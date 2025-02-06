package internal

import (
	"fmt"
	"slices"
	"sort"
	"sync"

	"github.com/mandelsoft/goutils/exception"
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/maputils"
	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/mandelsoft/goutils/stringutils"
)

type CredentialRecursion []ConsumerIdentity

func (c CredentialRecursion) String() string {
	return stringutils.Join(c)
}

func (c CredentialRecursion) Contains(identity ConsumerIdentity) bool {
	return slices.ContainsFunc(c, general.ContainsFuncFor(identity))
}

func (c CredentialRecursion) Append(identity ConsumerIdentity) CredentialRecursion {
	return sliceutils.CopyAppendUniqueFunc(c, general.EqualsFuncFor[ConsumerIdentity](), identity)
}

// UsageContext describes a dedicated type specific
// sub usage kinds for an object requiring credentials.
// For example, for an object providing a hierarchical
// namespace this might be a namespace prefix for
// included objects, for which credentials should be requested.
type UsageContext interface {
	String() string
}

type StringUsageContext string

func (s StringUsageContext) String() string {
	return string(s)
}

// ConsumerIdentityProvider is an interface for objects requiring
// credentials, which want to expose the ConsumerId they are
// using to request implicit credentials.
type ConsumerIdentityProvider interface {
	// GetConsumerId provides information about the consumer id
	// used for the object implementing this interface.
	// Optionally a sub context can be given to specify
	// a dedicated type specific sub realm.
	GetConsumerId(uctx ...UsageContext) ConsumerIdentity
	// GetIdentityMatcher provides the identity macher type to use
	// to match the consumer identities configured in a credentials
	// context.
	GetIdentityMatcher() string
}

type _consumers struct {
	data map[string]*_consumer
}

func newConsumers() *_consumers {
	return &_consumers{
		data: map[string]*_consumer{},
	}
}

func (c *_consumers) Set(id ConsumerIdentity, pid ProviderIdentity, creds CredentialsSource) {
	c.data[string(id.Key())] = &_consumer{
		providerId:  pid,
		identity:    id,
		credentials: creds,
	}
}

func (p *_consumers) Unregister(pid ProviderIdentity) {
	for n, c := range p.data {
		if c.providerId == pid {
			delete(p.data, n)
		}
	}
}

func (c *_consumers) Get(id ConsumerIdentity) (CredentialsSource, bool) {
	cred, ok := c.data[string(id.Key())]
	if cred != nil {
		return cred.credentials, true
	}
	return nil, ok
}

// Match matches a given request (pattern) against configured
// identities.
func (c *_consumers) Match(ectx EvaluationContext, pattern ConsumerIdentity, cur ConsumerIdentity, m IdentityMatcher) (CredentialsSource, ConsumerIdentity) {
	var found *_consumer
	for _, s := range c.data {
		if m(pattern, cur, s.identity) {
			found = s
			cur = s.identity
		}
	}
	if found != nil {
		return found.credentials, cur
	}
	return nil, cur
}

type _consumer struct {
	providerId  ProviderIdentity
	identity    ConsumerIdentity
	credentials CredentialsSource
}

func (c *_consumer) GetCredentials() CredentialsSource {
	return c.credentials
}

// //////////////////////////////////////////////////////////////////////////////

type consumerPrio struct {
	ConsumerProvider
	priority int
}

func (c *consumerPrio) GetPriority() int {
	return c.priority
}

func WithPriority(p ConsumerProvider, prio int) ConsumerProvider {
	return &consumerPrio{
		p,
		prio,
	}
}

// //////////////////////////////////////////////////////////////////////////////

type PriorityProvider interface {
	GetPriority() int
}

func priority(p interface{}) int {
	if pp, ok := p.(PriorityProvider); ok {
		return pp.GetPriority()
	}
	return 10
}

type consumerProviderRegistry struct {
	lock      sync.RWMutex
	explicit  *_consumers
	providers map[ProviderIdentity]ConsumerProvider
	ordered   []ConsumerProvider
}

func newConsumerProviderRegistry() *consumerProviderRegistry {
	return &consumerProviderRegistry{
		explicit:  newConsumers(),
		providers: map[ProviderIdentity]ConsumerProvider{},
		ordered:   nil,
	}
}

var _ ConsumerProvider = (*consumerProviderRegistry)(nil)

func (p *consumerProviderRegistry) Register(id ProviderIdentity, c ConsumerProvider) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.unregister(id)
	p.providers[id] = c
	p.ordered = maputils.OrderedValues(p.providers)
	sort.Slice(p.ordered, func(a, b int) bool {
		return priority(p.ordered[a]) < priority(p.ordered[b])
	})
}

func (p *consumerProviderRegistry) Unregister(id ProviderIdentity) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.unregister(id)
}

func (p *consumerProviderRegistry) unregister(id ProviderIdentity) {
	p.explicit.Unregister(id)
	if _, ok := p.providers[id]; ok {
		delete(p.providers, id)
		p.ordered = maputils.OrderedValues(p.providers)
		sort.Slice(p.ordered, func(a, b int) bool {
			return priority(p.ordered[a]) < priority(p.ordered[b])
		})
	} else {
		for _, sub := range p.providers {
			sub.Unregister(id)
		}
	}
}

func (p *consumerProviderRegistry) Get(id ConsumerIdentity) (CredentialsSource, bool) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	credsrc, ok := p.explicit.Get(id)
	if ok {
		return credsrc, ok
	}
	for _, sub := range p.providers {
		credsrc, ok := sub.Get(id)
		if ok {
			return credsrc, ok
		}
	}
	return nil, false
}

func (p *consumerProviderRegistry) checkHandleProvider(ectx EvaluationContext, prov ConsumerProvider, pattern ConsumerIdentity) (rctx EvaluationContext, useprov bool, usestack bool) {
	if pr, ok := prov.(ConsumerIdentityProvider); ok {
		r := GetEvaluationContextFor[CredentialRecursion](ectx)
		if r == nil {
			r = CredentialRecursion{}
		}
		if r.Contains(pr.GetConsumerId()) {
			return ectx, false, true
		}
		r = r.Append(pr.GetConsumerId())
		ectx = SetEvaluationContextFor(ectx, r)
	}
	return ectx, true, true
}

type UnwindStack struct {
	error
}

func (u *UnwindStack) Unwrap() error {
	return u.error
}

func (p *consumerProviderRegistry) catchedMatch(ectx EvaluationContext, sub ConsumerProvider, pattern ConsumerIdentity, cur ConsumerIdentity, m IdentityMatcher) (cs CredentialsSource, ci ConsumerIdentity) {
	defer exception.CatchError(func(err error) {
		log.Trace("caught unwind stack error: {{error}}", "error", err)
		cs = nil
		ci = cur
	}, exception.ByPrototypes(&UnwindStack{}))
	log.Trace("pattern: {{pattern}}\ncontext: {{context}}",
		"pattern", pattern, "context", ectx)
	ectx, useprov, _ := p.checkHandleProvider(ectx, sub, pattern)
	if !useprov {
		return nil, cur
	}
	log.Trace("attempt match with provider")
	return sub.Match(ectx, pattern, cur, m)
}

func (p *consumerProviderRegistry) Match(ectx EvaluationContext, pattern ConsumerIdentity, cur ConsumerIdentity, m IdentityMatcher) (CredentialsSource, ConsumerIdentity) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	credsrc, cur := p.explicit.Match(ectx, pattern, cur, m)
	for _, sub := range p.providers {
		var f CredentialsSource
		f, cur = p.catchedMatch(ectx, sub, pattern, cur, m)
		if f != nil {
			credsrc = f
		}
	}
	// If this is the case, we are in a situation where we have excluded all providers (since they are all in the stack).
	// If we would simply return with no credentials, the follow-up coding would assume, that it should query the
	// credential repository without any credentials, since none have been found.
	// INSTEAD, we have to step down to the previous recursion level and check the other potentially available providers
	// for credentials.
	// BUT in case we have explicit credentials, then we should use those.
	if credsrc == nil {
		r := GetEvaluationContextFor[CredentialRecursion](ectx)
		// unwind the stack only makes sense when we are in a recursive call, thus we have at least one provider on the
		// credential recursion stack
		if len(r) > 0 && len(r) == len(p.providers) {
			exception.Throw(&UnwindStack{fmt.Errorf("impossible credential recursion detected - unwind stack")})
		}
	}
	log.Trace("return credential source")
	return credsrc, cur
}

func (p *consumerProviderRegistry) Set(id ConsumerIdentity, pid ProviderIdentity, creds CredentialsSource) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.explicit.Set(id, pid, creds)
}
