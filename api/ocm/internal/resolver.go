package internal

import (
	"strings"
	"sync"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/finalizer"
	"github.com/mandelsoft/goutils/general"
	"golang.org/x/exp/slices"

	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/utils"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/refmgmt"
	"ocm.software/ocm/api/utils/registrations"
)

////////////////////////////////////////////////////////////////////////////////

type ResolverRule struct {
	prefix string
	path   registrations.NamePath
	spec   RepositorySpec
	prio   int
}

func (r *ResolverRule) GetPrefix() string {
	return r.prefix
}

func (r *ResolverRule) GetSpecification() RepositorySpec {
	return r.spec
}

func (r *ResolverRule) GetPriority() int {
	return r.prio
}

// RepositoryCache is a utility object intended to be used by higher level objects such as session or resolver. Since
// the closing of the repository objects depends on the usage context (e.g. if components have been looked up in this
// repository, these components have to be closed before the repository can be closed), it is the responsibility of the
// higher level objects to close the repositories correctly.
type RepositoryCache struct {
	lock         sync.Mutex
	repositories map[datacontext.ObjectKey]Repository
}

func NewRepositoryCache() *RepositoryCache {
	return &RepositoryCache{
		repositories: map[datacontext.ObjectKey]Repository{},
	}
}

func (c *RepositoryCache) Reset() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.repositories = map[datacontext.ObjectKey]Repository{}
}

func (c *RepositoryCache) LookupRepository(ctx Context, spec RepositorySpec) (Repository, bool, error) {
	spec, err := ctx.RepositoryTypes().Convert(spec)
	if err != nil {
		return nil, false, err
	}
	keyName, err := utils.Key(spec)
	if err != nil {
		return nil, false, err
	}
	key := datacontext.ObjectKey{
		Object: ctx,
		Name:   keyName,
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	if r := c.repositories[key]; r != nil {
		return r, true, nil
	}
	repo, err := ctx.RepositoryForSpec(spec)
	if err != nil {
		return nil, false, err
	}
	c.repositories[key] = repo
	return repo, false, err
}

func NewResolverRule(prefix string, spec RepositorySpec, prio ...int) *ResolverRule {
	p := registrations.NewNamePath(prefix)
	return &ResolverRule{
		prefix: prefix,
		path:   p,
		spec:   spec,
		prio:   general.OptionalDefaulted(10, prio...),
	}
}

func (r *ResolverRule) Compare(o *ResolverRule) int {
	if d := r.prio - o.prio; d != 0 {
		return d
	}
	return r.path.Compare(o.path)
}

func (r *ResolverRule) Match(name string) bool {
	return r.prefix == "" || r.prefix == name || strings.HasPrefix(name, r.prefix+"/")
}

// MatchingResolver hosts rule to match component version names.
// Matched names will be mapped to a specification for repository
// which should be used to look up the component version.
// Therefore, it keeps a reference to the context to use.
//
// ATTENTION: Because such an object is used by the context
// implementation, the context must be kept as ContextProvider
// to provide context views to outbound calls.
type MatchingResolver struct {
	lock     sync.Mutex
	ctx      ContextProvider
	finalize finalizer.Finalizer
	cache    *RepositoryCache
	rules    []*ResolverRule
}

func NewMatchingResolver(ctx ContextProvider, rules ...*ResolverRule) *MatchingResolver {
	return &MatchingResolver{
		lock:  sync.Mutex{},
		ctx:   ctx,
		cache: NewRepositoryCache(),
		rules: nil,
	}
}

func (r *MatchingResolver) OCMContext() Context {
	return r.ctx.OCMContext()
}

func (r *MatchingResolver) Finalize() error {
	r.lock.Lock()
	defer r.lock.Unlock()
	defer r.cache.Reset()
	return r.finalize.Finalize()
}

func (r *MatchingResolver) GetRules() []*ResolverRule {
	r.lock.Lock()
	defer r.lock.Unlock()
	return slices.Clone(r.rules)
}

func (r *MatchingResolver) AddRule(prefix string, spec RepositorySpec, prio ...int) {
	r.lock.Lock()
	defer r.lock.Unlock()

	rule := NewResolverRule(prefix, spec, prio...)
	found := len(r.rules)
	for i, o := range r.rules {
		if o.Compare(rule) < 0 {
			found = i
			break
		}
	}
	r.rules = slices.Insert(r.rules, found, rule)
}

func (r *MatchingResolver) LookupComponentVersion(name string, version string) (ComponentVersionAccess, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	ctx := r.ctx.OCMContext()
	for _, rule := range r.rules {
		if rule.Match(name) {
			repo, cached, err := r.cache.LookupRepository(ctx, rule.spec)
			if err != nil {
				return nil, err
			}
			if !cached {
				// Even though the matching resolver is closed, there might be components or component versions, which
				// contain a reference to the repository. Still, it shall be possible to close the matching resolver.
				refmgmt.Lazy(repo)
				r.finalize.Close(repo)
			}
			cv, err := repo.LookupComponentVersion(name, version)
			if err == nil && cv != nil {
				return cv, nil
			}
			if !errors.IsErrNotFoundKind(err, KIND_COMPONENTVERSION) {
				return nil, err
			}
		}
	}
	return nil, errors.ErrNotFound(KIND_COMPONENTVERSION, common.NewNameVersion(name, version).String())
}
