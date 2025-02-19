package internal

import (
	"slices"
	"strings"
	"sync"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/finalizer"
	"github.com/mandelsoft/goutils/general"

	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/utils"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/refmgmt"
	"ocm.software/ocm/api/utils/registrations"
)

////////////////////////////////////////////////////////////////////////////////

// ComponentVersionResolver describes an object able
// to map component version identities to a component version access,
// the representtaion of the component version stored in a dedicated
// OCM repository.
// Such a resolver might optionally implement the ComponentResolver
// interface to provide information about potential locations for
// a particular component.
type ComponentVersionResolver interface {
	LookupComponentVersion(name string, version string) (ComponentVersionAccess, error)
}

// ResolvedComponentProvider is an interface for an
// object providing optional access to a repository
// and a provider for component versions.
// Direct operations and operations of provided objects
// may fail, if an optional underlying object is already closed
// Therefore, it does not need to be explicitly closed.
//
// Be careful: Objects of this plain type should only be held as long as the
// providing object is valid and not closed.
type ResolvedComponentProvider interface {
	Repository() (Repository, error)
	LookupComponent(name string) (ResolvedComponentVersionProvider, error)
}

// ResolvedComponentVersionProvider is an interface for an
// object providing access component versions.
// Operations may fail, if an underlying object is already closed.
//
// Be careful: Objects of this plain type should only be held as long as the
// providing object is valid and not closed.
type ResolvedComponentVersionProvider interface {
	GetName() string
	LookupVersion(version string) (ComponentVersionAccess, error)
	ListVersions() ([]string, error)
}

// ComponentResolver is an optional interface for a ComponentVersionResolver,
// which can offer information about potential providers usable to lookup
// versions for a dedicated component.
//
// Be careful, provided objects may refer to closable objects
// held by the interface provider. They should only be used as long
// as the object behind the ComponentResolver is not closed.
type ComponentResolver interface {
	LookupComponentProviders(name string) []ResolvedComponentProvider
}

////////////////////////////////////////////////////////////////////////////////

type componentProvider struct {
	repo Repository
}

func RepositoryProviderForRepository(r Repository) ResolvedComponentProvider {
	return &componentProvider{r}
}

func (p *componentProvider) Repository() (Repository, error) {
	return p.repo.Dup()
}

func (p *componentProvider) LookupComponent(name string) (ResolvedComponentVersionProvider, error) {
	return &versionProvider{name, p}, nil
}

type versionProvider struct {
	name string
	repo *componentProvider
}

func (p *versionProvider) GetName() string {
	return p.name
}

func (p *versionProvider) lookupComponent() (ComponentAccess, error) {
	r, err := refmgmt.ToLazy(p.repo.Repository())
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return r.LookupComponent(p.name)
}

func (p *versionProvider) LookupVersion(version string) (ComponentVersionAccess, error) {
	c, err := refmgmt.ToLazy(p.lookupComponent())
	if err != nil {
		return nil, err
	}
	defer c.Close()
	return c.LookupVersion(version)
}

func (p *versionProvider) ListVersions() ([]string, error) {
	c, err := p.lookupComponent()
	if err != nil {
		return nil, err
	}
	defer c.Close()
	return c.ListVersions()
}

////////////////////////////////////////////////////////////////////////////////

type ResolverRule interface {
	GetPrefix() string
	GetPath() registrations.NamePath
	GetSpecification() RepositorySpec
	GetPriority() int

	Match(name string) bool
}

type resolverRule struct {
	prefix string
	path   registrations.NamePath
	spec   RepositorySpec
	prio   int
}

func (r *resolverRule) GetPrefix() string {
	return r.prefix
}

func (r *resolverRule) GetPath() registrations.NamePath {
	return slices.Clone(r.path)
}

func (r *resolverRule) GetSpecification() RepositorySpec {
	return r.spec
}

func (r *resolverRule) GetPriority() int {
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

func CompareRule(r, o ResolverRule) int {
	if d := r.GetPriority() - o.GetPriority(); d != 0 {
		return d
	}
	return r.GetPath().Compare(o.GetPath())
}

func NewResolverRule(prefix string, spec RepositorySpec, prio ...int) ResolverRule {
	p := registrations.NewNamePath(prefix)
	return &resolverRule{
		prefix: prefix,
		path:   p,
		spec:   spec,
		prio:   general.OptionalDefaulted(10, prio...),
	}
}

func (r *resolverRule) Match(name string) bool {
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
	rules    []ResolverRule
}

var _ ComponentResolver = (*MatchingResolver)(nil)

func NewMatchingResolver(ctx ContextProvider, rules ...ResolverRule) *MatchingResolver {
	return &MatchingResolver{
		lock:  sync.Mutex{},
		ctx:   ctx,
		cache: NewRepositoryCache(),
		rules: slices.Clone(rules),
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

func (r *MatchingResolver) HasRules() bool {
	r.lock.Lock()
	defer r.lock.Unlock()
	return len(r.rules) != 0
}

func (r *MatchingResolver) GetRules() []ResolverRule {
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
		if CompareRule(o, rule) < 0 {
			found = i
			break
		}
	}
	r.rules = slices.Insert(r.rules, found, rule)
}

func (r *MatchingResolver) LookupComponentVersion(name string, version string) (ComponentVersionAccess, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	for _, rule := range r.rules {
		if rule.Match(name) {
			repo, err := r.resolveRepository(rule)
			if err != nil {
				return nil, err
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

func (r *MatchingResolver) LookupComponentProviders(name string) []ResolvedComponentProvider {
	r.lock.Lock()
	defer r.lock.Unlock()

	var result []ResolvedComponentProvider
	for _, rule := range r.rules {
		if rule.Match(name) {
			result = append(result, &cachedRepository{r, rule})
		}
	}
	return result
}

func (r *MatchingResolver) resolveRepository(rule ResolverRule) (Repository, error) {
	repo, cached, err := r.cache.LookupRepository(r.ctx.OCMContext(), rule.GetSpecification())
	if err != nil {
		return nil, err
	}
	if !cached {
		// Even though the matching resolver is closed, there might be components or component versions, which
		// contain a reference to the repository. Still, it shall be possible to close the matching resolver.
		refmgmt.Lazy(repo)
		r.finalize.Close(repo)
	}
	return repo, nil
}

type cachedRepository struct {
	resolver *MatchingResolver
	rule     ResolverRule
}

func (c *cachedRepository) Repository() (Repository, error) {
	c.resolver.lock.Lock()
	defer c.resolver.lock.Unlock()

	repo, err := c.resolver.resolveRepository(c.rule)
	if err != nil {
		return nil, err
	}
	return repo.Dup()
}

func (c *cachedRepository) LookupComponent(name string) (ResolvedComponentVersionProvider, error) {
	return &cachedComponent{name, c}, nil
}

type cachedComponent struct {
	name string
	*cachedRepository
}

func (c *cachedComponent) GetName() string {
	return c.name
}

func (c *cachedComponent) LookupVersion(version string) (ComponentVersionAccess, error) {
	c.resolver.lock.Lock()
	defer c.resolver.lock.Unlock()

	repo, err := c.resolver.resolveRepository(c.rule)
	if err != nil {
		return nil, err
	}
	return repo.LookupComponentVersion(c.name, version)
}

func (c *cachedComponent) ListVersions() ([]string, error) {
	c.resolver.lock.Lock()
	defer c.resolver.lock.Unlock()

	repo, err := c.resolver.resolveRepository(c.rule)
	if err != nil {
		return nil, err
	}
	ca, err := repo.LookupComponent(c.name)
	if err != nil {
		return nil, err
	}
	defer ca.Close()
	return ca.ListVersions()
}
