package resolvers

import (
	"sync"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/sliceutils"

	"ocm.software/ocm/api/ocm/internal"
	"ocm.software/ocm/api/utils/semverutils"
)

type componentResolver struct {
	ResolvedComponentProvider
}

var _ ComponentResolver = (*componentResolver)(nil)

func ComponentResolverForRepository(repo Repository) ComponentResolver {
	return &componentResolver{internal.RepositoryProviderForRepository(repo)}
}

func (c componentResolver) LookupComponentProviders(name string) []ResolvedComponentProvider {
	return []ResolvedComponentProvider{c}
}

////////////////////////////////////////////////////////////////////////////////

type CompoundComponentResolver struct {
	lock      sync.RWMutex
	resolvers []ComponentResolver
}

var _ ComponentResolver = (*CompoundComponentResolver)(nil)

func NewCompoundComponentResolver(res ...ComponentResolver) ComponentResolver {
	for i := 0; i < len(res); i++ {
		if res[i] == nil {
			res = append(res[:i], res[i+1:]...)
			i--
		}
	}
	if len(res) == 1 {
		return res[0]
	}
	return &CompoundComponentResolver{resolvers: res}
}

func (c *CompoundComponentResolver) LookupComponentProviders(name string) []ResolvedComponentProvider {
	c.lock.RLock()
	defer c.lock.RUnlock()

	var list []ResolvedComponentProvider
	for _, r := range c.resolvers {
		if r == nil {
			continue
		}

		r.LookupComponentProviders(name)
		list = append(list, r.LookupComponentProviders(name)...)
	}
	return list
}

func (c *CompoundComponentResolver) AddResolver(r ComponentResolver) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if r != nil {
		c.resolvers = append(c.resolvers, r)
	}
}

////////////////////////////////////////////////////////////////////////////////

type cvrForCr struct {
	resolver ComponentResolver
}

func ComponentVersionResolverForComponentResolver(r ComponentResolver) ComponentVersionResolver {
	return &cvrForCr{r}
}

func (c *cvrForCr) LookupComponentVersion(comp string, vers string) (ComponentVersionAccess, error) {
	return LookupComponentVersion(comp, vers, c.resolver)
}

////////////////////////////////////////////////////////////////////////////////

func ListComponentVersions(comp string, r ComponentResolver) ([]string, error) {
	var (
		versions []string
		errlist  errors.ErrorList
	)

	for _, p := range r.LookupComponentProviders(comp) {
		c, err := p.LookupComponent(comp)
		if err != nil || c == nil {
			if !errors.IsErrNotFound(err) {
				errlist.Add(err)
			}
			continue
		}
		list, err := c.ListVersions()
		if err != nil {
			errlist.Add(err)
			continue
		}
		versions = sliceutils.AppendUnique(versions, list...)
	}
	semverutils.SortVersions(versions)
	return versions, errlist.Result()
}

func LookupComponentVersion(comp, vers string, r ComponentResolver) (ComponentVersionAccess, error) {
	var errlist errors.ErrorList

	for _, p := range r.LookupComponentProviders(comp) {
		c, err := p.LookupComponent(comp)
		if err != nil || c == nil {
			if !errors.IsErrNotFound(err) {
				errlist.Add(err)
			}
			continue
		}

		cv, err := c.LookupVersion(vers)
		if err != nil || cv == nil {
			if !errors.IsErrNotFound(err) {
				errlist.Add(err)
			}
			continue
		}
		return cv, nil
	}
	return nil, errlist.Result()
}
