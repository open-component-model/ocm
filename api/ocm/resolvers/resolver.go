package resolvers

import (
	"slices"
	"sync"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/sliceutils"

	"ocm.software/ocm/api/ocm/internal"
	common "ocm.software/ocm/api/utils/misc"
)

type DedicatedResolver []ComponentVersionAccess

var (
	_ ComponentVersionResolver = (*DedicatedResolver)(nil)
	_ ComponentResolver        = (*DedicatedResolver)(nil)
)

func NewDedicatedResolver(cv ...ComponentVersionAccess) ComponentVersionResolver {
	return DedicatedResolver(slices.Clone(cv))
}

func (d DedicatedResolver) Repository() (Repository, error) {
	return nil, nil
}

func (d DedicatedResolver) LookupComponentVersion(name string, version string) (ComponentVersionAccess, error) {
	for _, cv := range d {
		if cv.GetName() == name && cv.GetVersion() == version {
			return cv.Dup()
		}
	}
	return nil, nil
}

func (d DedicatedResolver) LookupComponentProviders(name string) []ResolvedComponentProvider {
	for _, c := range d {
		if c.GetName() == name {
			return []ResolvedComponentProvider{d}
		}
	}
	return nil
}

func (d DedicatedResolver) LookupComponent(name string) (ResolvedComponentVersionProvider, error) {
	return &versionProvider{name, d}, nil
}

type versionProvider struct {
	name     string
	resolver DedicatedResolver
}

func (p *versionProvider) GetName() string {
	return p.name
}

func (p *versionProvider) LookupVersion(vers string) (ComponentVersionAccess, error) {
	return p.resolver.LookupComponentVersion(p.name, vers)
}

func (p *versionProvider) ListVersions() ([]string, error) {
	var vers []string
	for _, c := range p.resolver {
		if c.GetName() == p.name {
			vers = sliceutils.AppendUnique(vers, c.GetVersion())
		}
	}
	return vers, nil
}

////////////////////////////////////////////////////////////////////////////////

type CompoundResolver struct {
	lock      sync.RWMutex
	resolvers []ComponentVersionResolver
}

var (
	_ ComponentVersionResolver = (*CompoundResolver)(nil)
	_ ComponentResolver        = (*CompoundResolver)(nil)
)

func NewCompoundResolver(res ...ComponentVersionResolver) ComponentVersionResolver {
	for i := 0; i < len(res); i++ {
		if res[i] == nil {
			res = append(res[:i], res[i+1:]...)
			i--
		}
	}
	if len(res) == 1 {
		return res[0]
	}
	return &CompoundResolver{resolvers: res}
}

func (c *CompoundResolver) LookupComponentVersion(name string, version string) (ComponentVersionAccess, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	for _, r := range c.resolvers {
		if r == nil {
			continue
		}
		cv, err := r.LookupComponentVersion(name, version)
		if err == nil && cv != nil {
			return cv, nil
		}
		if !errors.IsErrNotFoundKind(err, KIND_COMPONENTVERSION) && !errors.IsErrNotFoundKind(err, KIND_COMPONENT) {
			return nil, err
		}
	}
	return nil, errors.ErrNotFound(KIND_OCM_REFERENCE, common.NewNameVersion(name, version).String())
}

func (c *CompoundResolver) LookupComponentProviders(name string) []ResolvedComponentProvider {
	c.lock.RLock()
	defer c.lock.RUnlock()

	var result []ResolvedComponentProvider

	for _, r := range c.resolvers {
		if cr, ok := r.(ComponentResolver); ok {
			result = append(result, cr.LookupComponentProviders(name)...)
		}
	}
	return result
}

func (c *CompoundResolver) AddResolver(r ComponentVersionResolver) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.resolvers = append(c.resolvers, r)
}

////////////////////////////////////////////////////////////////////////////////

type MatchingResolver interface {
	ComponentVersionResolver
	ContextProvider

	AddRule(prefix string, spec RepositorySpec, prio ...int)
	Finalize() error
}

func NewMatchingResolver(ctx ContextProvider, rules ...ResolverRule) MatchingResolver {
	return internal.NewMatchingResolver(ctx.OCMContext(), rules...)
}
