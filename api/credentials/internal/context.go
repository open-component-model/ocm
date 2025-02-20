package internal

import (
	"context"
	"fmt"
	"maps"
	"reflect"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/goutils/maputils"

	"ocm.software/ocm/api/config"
	cfgcpi "ocm.software/ocm/api/config/cpi"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/api/utils/runtimefinalizer"
)

// CONTEXT_TYPE is the global type for a credential context.
const CONTEXT_TYPE = "credentials" + datacontext.OCM_CONTEXT_SUFFIX

// ProviderIdentity is used to uniquely identify a provider
// for a configured consumer id. If non-empty it
// must start with a DNSname identifying the origin of the
// provider followed by a slash and a local arbitrary identity.
type ProviderIdentity = runtimefinalizer.ObjectIdentity

type ContextProvider interface {
	CredentialsContext() Context
}

type ConsumerProvider interface {
	Unregister(id ProviderIdentity)
	Get(id ConsumerIdentity) (CredentialsSource, bool)
	Match(ectx EvaluationContext, id ConsumerIdentity, cur ConsumerIdentity, matcher IdentityMatcher) (CredentialsSource, ConsumerIdentity)
}

type EvaluationContext *evaluationContext

type evaluationContext struct {
	data map[reflect.Type]interface{}
}

func (e evaluationContext) String() string {
	return fmt.Sprintf("%v", maputils.Transform(e.data, func(k reflect.Type, v interface{}) (string, string) {
		return k.Name(), fmt.Sprintf("%v", v)
	}))
}

func GetEvaluationContextFor[T any](ectx EvaluationContext) T {
	var _nil T
	if ectx.data == nil {
		return _nil
	}
	return generics.Cast[T](ectx.data[generics.TypeOf[T]()])
}

func SetEvaluationContextFor(ectx EvaluationContext, e any) EvaluationContext {
	if ectx.data == nil {
		ectx.data = map[reflect.Type]interface{}{}
	}
	n := &evaluationContext{maps.Clone(ectx.data)}
	n.data[reflect.TypeOf(e)] = e
	return n
}

type Context interface {
	datacontext.Context
	ContextProvider
	config.ContextProvider

	AttributesContext() datacontext.AttributesContext
	RepositoryTypes() RepositoryTypeScheme

	RepositorySpecForConfig(data []byte, unmarshaler runtime.Unmarshaler) (RepositorySpec, error)

	RepositoryForSpec(spec RepositorySpec, creds ...CredentialsSource) (Repository, error)
	RepositoryForConfig(data []byte, unmarshaler runtime.Unmarshaler, creds ...CredentialsSource) (Repository, error)

	CredentialsForSpec(spec CredentialsSpec, creds ...CredentialsSource) (Credentials, error)
	CredentialsForConfig(data []byte, unmarshaler runtime.Unmarshaler, cred ...CredentialsSource) (Credentials, error)

	RegisterConsumerProvider(id ProviderIdentity, provider ConsumerProvider)
	UnregisterConsumerProvider(id ProviderIdentity)

	GetCredentialsForConsumer(ConsumerIdentity, ...IdentityMatcher) (CredentialsSource, error)
	getCredentialsForConsumer(EvaluationContext, ConsumerIdentity, ...IdentityMatcher) (CredentialsSource, error)
	SetCredentialsForConsumer(identity ConsumerIdentity, creds CredentialsSource)
	SetCredentialsForConsumerWithProvider(pid ProviderIdentity, identity ConsumerIdentity, creds CredentialsSource)

	SetAlias(name string, spec RepositorySpec, creds ...CredentialsSource) error

	ConsumerIdentityMatchers() IdentityMatcherRegistry
}

var key = reflect.TypeOf(_context{})

// DefaultContext is the default context initialized by init functions.
var DefaultContext = Builder{}.New(datacontext.MODE_SHARED)

// FromContext returns the Context to use for context.Context.
// This is either an explicit context or the default context.
func FromContext(ctx context.Context) Context {
	c, _ := datacontext.ForContextByKey(ctx, key, DefaultContext)
	return c.(Context)
}

func FromProvider(p ContextProvider) Context {
	if p == nil {
		return nil
	}
	return p.CredentialsContext()
}

func DefinedForContext(ctx context.Context) (Context, bool) {
	c, ok := datacontext.ForContextByKey(ctx, key, DefaultContext)
	if c != nil {
		return c.(Context), ok
	}
	return nil, ok
}

type _InternalContext = datacontext.InternalContext

type _context struct {
	_InternalContext

	sharedattributes         datacontext.AttributesContext
	updater                  cfgcpi.Updater
	knownRepositoryTypes     RepositoryTypeScheme
	consumerIdentityMatchers IdentityMatcherRegistry
	consumerProviders        *consumerProviderRegistry
}

var (
	_ Context                          = (*_context)(nil)
	_ datacontext.ViewCreator[Context] = (*_context)(nil)
)

// gcWrapper is used as garbage collectable
// wrapper for a context implementation
// to establish a runtime finalizer.
type gcWrapper struct {
	datacontext.GCWrapper
	*_context
}

func newView(c *_context, ref ...bool) Context {
	if utils.Optional(ref...) {
		return datacontext.FinalizedContext[gcWrapper](c)
	}
	return c
}

func (w *gcWrapper) SetContext(c *_context) {
	w._context = c
}

func newContext(configctx config.Context, reposcheme RepositoryTypeScheme, consumerMatchers IdentityMatcherRegistry, delegates datacontext.Delegates) Context {
	c := &_context{
		sharedattributes:         datacontext.PersistentContextRef(configctx.AttributesContext()),
		knownRepositoryTypes:     reposcheme,
		consumerIdentityMatchers: consumerMatchers,
		consumerProviders:        newConsumerProviderRegistry(),
	}
	c._InternalContext = datacontext.NewContextBase(c, CONTEXT_TYPE, key, configctx.GetAttributes(), delegates)
	c.updater = cfgcpi.NewUpdaterForFactory(datacontext.PersistentContextRef(configctx), c.CredentialsContext)
	return newView(c, true)
}

func (c *_context) CreateView() Context {
	return newView(c, true)
}

func (c *_context) CredentialsContext() Context {
	return newView(c)
}

func (c *_context) Update() error {
	return c.updater.Update()
}

func (c *_context) GetType() string {
	return CONTEXT_TYPE
}

func (c *_context) AttributesContext() datacontext.AttributesContext {
	return c.sharedattributes
}

func (c *_context) ConfigContext() config.Context {
	return c.updater.GetContext()
}

func (c *_context) RepositoryTypes() RepositoryTypeScheme {
	return c.knownRepositoryTypes
}

func (c *_context) RepositorySpecForConfig(data []byte, unmarshaler runtime.Unmarshaler) (RepositorySpec, error) {
	return c.knownRepositoryTypes.Decode(data, unmarshaler)
}

func (c *_context) RepositoryForSpec(spec RepositorySpec, creds ...CredentialsSource) (Repository, error) {
	out := newView(c)
	cred, err := CredentialsChain(creds).Credentials(out)
	if err != nil {
		return nil, err
	}
	c.Update()
	return spec.Repository(out, cred)
}

func (c *_context) RepositoryForConfig(data []byte, unmarshaler runtime.Unmarshaler, creds ...CredentialsSource) (Repository, error) {
	spec, err := c.knownRepositoryTypes.Decode(data, unmarshaler)
	if err != nil {
		return nil, err
	}
	return c.RepositoryForSpec(spec, creds...)
}

func (c *_context) CredentialsForSpec(spec CredentialsSpec, creds ...CredentialsSource) (Credentials, error) {
	out := newView(c)
	repospec := spec.GetRepositorySpec(out)
	repo, err := c.RepositoryForSpec(repospec, creds...)
	if err != nil {
		return nil, err
	}
	return repo.LookupCredentials(spec.GetCredentialsName())
}

func (c *_context) CredentialsForConfig(data []byte, unmarshaler runtime.Unmarshaler, creds ...CredentialsSource) (Credentials, error) {
	spec := &GenericCredentialsSpec{}
	err := unmarshaler.Unmarshal(data, spec)
	if err != nil {
		return nil, err
	}
	return c.CredentialsForSpec(spec, creds...)
}

var emptyIdentity = ConsumerIdentity{}

func (c *_context) GetCredentialsForConsumer(identity ConsumerIdentity, matchers ...IdentityMatcher) (CredentialsSource, error) {
	return c.getCredentialsForConsumer(nil, identity, matchers...)
}

func (c *_context) getCredentialsForConsumer(ectx EvaluationContext, identity ConsumerIdentity, matchers ...IdentityMatcher) (CredentialsSource, error) {
	err := c.Update()
	if err != nil {
		return nil, err
	}

	if ectx == nil {
		ectx = &evaluationContext{}
	}
	m := c.defaultMatcher(identity, matchers...)
	var credsrc CredentialsSource
	if m == nil {
		credsrc, _ = c.consumerProviders.Get(identity)
	} else {
		credsrc, _ = c.consumerProviders.Match(ectx, identity, nil, m)
	}
	if credsrc == nil {
		credsrc, _ = c.consumerProviders.Get(emptyIdentity)
	}
	if credsrc == nil {
		return nil, ErrUnknownConsumer(identity.String())
	}
	return credsrc, nil
}

func (c *_context) defaultMatcher(id ConsumerIdentity, matchers ...IdentityMatcher) IdentityMatcher {
	def := c.consumerIdentityMatchers.Get(id.Type())
	if def == nil {
		def = PartialMatch
	}
	return mergeMatcher(def, andMatcher, matchers)
}

func (c *_context) SetCredentialsForConsumer(identity ConsumerIdentity, creds CredentialsSource) {
	c.Update()
	c.consumerProviders.Set(identity, "", creds)
}

func (c *_context) SetCredentialsForConsumerWithProvider(pid ProviderIdentity, identity ConsumerIdentity, creds CredentialsSource) {
	c.Update()
	c.consumerProviders.Set(identity, pid, creds)
}

func (c *_context) ConsumerIdentityMatchers() IdentityMatcherRegistry {
	return c.consumerIdentityMatchers
}

func (c *_context) SetAlias(name string, spec RepositorySpec, creds ...CredentialsSource) error {
	c.Update()
	t := c.knownRepositoryTypes.GetType(AliasRepositoryType)
	if t == nil {
		return errors.ErrNotSupported("aliases")
	}
	if a, ok := t.(AliasRegistry); ok {
		return a.SetAlias(c, name, spec, CredentialsChain(creds))
	}
	return errors.ErrNotImplemented("interface", "AliasRegistry", reflect.TypeOf(t).String())
}

func (c *_context) RegisterConsumerProvider(id ProviderIdentity, provider ConsumerProvider) {
	c.consumerProviders.Register(id, provider)
}

func (c *_context) UnregisterConsumerProvider(id ProviderIdentity) {
	c.consumerProviders.Unregister(id)
}

///////////////////////////////////////

func GetCredentialsForConsumer(ctx Context, ectx EvaluationContext, identity ConsumerIdentity, matchers ...IdentityMatcher) (CredentialsSource, error) {
	if ectx == nil {
		ectx = &evaluationContext{}
	}
	return ctx.getCredentialsForConsumer(ectx, identity, matchers...)
}
