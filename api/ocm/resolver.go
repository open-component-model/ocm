package ocm

import (
	"sync"

	"github.com/mandelsoft/goutils/errors"
	"golang.org/x/exp/slices"

	"ocm.software/ocm/api/ocm/internal"
	common "ocm.software/ocm/api/utils/misc"
)

type DedicatedResolver []ComponentVersionAccess

var _ ComponentVersionResolver = (*DedicatedResolver)(nil)

func NewDedicatedResolver(cv ...ComponentVersionAccess) ComponentVersionResolver {
	return DedicatedResolver(slices.Clone(cv))
}

func (d DedicatedResolver) LookupComponentVersion(name string, version string) (ComponentVersionAccess, error) {
	for _, cv := range d {
		if cv.GetName() == name && cv.GetVersion() == version {
			return cv.Dup()
		}
	}
	return nil, nil
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

func (c *CompoundResolver) LookupComponentVersion(name string, version string) (internal.ComponentVersionAccess, error) {
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

func (c *CompoundResolver) LookupRepositoriesForComponent(name string) []internal.RepositoryProvider {
	c.lock.RLock()
	defer c.lock.RUnlock()

	var result []RepositoryProvider

	for _, r := range c.resolvers {
		if cr, ok := r.(ComponentResolver); ok {
			result = append(result, cr.LookupRepositoriesForComponent(name)...)
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

func NewMatchingResolver(ctx ContextProvider) MatchingResolver {
	return internal.NewMatchingResolver(ctx.OCMContext())
}
