package compdesc

import (
	"sync"

	"github.com/mandelsoft/goutils/errors"
	common "ocm.software/ocm/api/utils/misc"
)

type ComponentVersionResolver interface {
	LookupComponentVersion(name string, version string) (*ComponentDescriptor, error)
}

////////////////////////////////////////////////////////////////////////////////

type ComponentVersionSet struct {
	lock sync.RWMutex
	cds  map[common.NameVersion]*ComponentDescriptor
}

var _ ComponentVersionResolver = (*ComponentVersionSet)(nil)

func NewComponentVersionSet(cds ...*ComponentDescriptor) *ComponentVersionSet {
	r := map[common.NameVersion]*ComponentDescriptor{}
	for _, cd := range cds {
		r[common.NewNameVersion(cd.Name, cd.Version)] = cd.Copy()
	}
	return &ComponentVersionSet{cds: r}
}

func (c *ComponentVersionSet) LookupComponentVersion(name string, version string) (*ComponentDescriptor, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	nv := common.NewNameVersion(name, version)

	cd := c.cds[nv]
	if cd == nil {
		return nil, errors.ErrNotFound(KIND_COMPONENTVERSION, nv.String())
	}
	return cd, nil
}

func (c *ComponentVersionSet) AddVersion(cd *ComponentDescriptor) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.cds[common.NewNameVersion(cd.Name, cd.Version)] = cd.Copy()
}

////////////////////////////////////////////////////////////////////////////////

type CompoundResolver struct {
	lock      sync.RWMutex
	resolvers []ComponentVersionResolver
}

var _ ComponentVersionResolver = (*CompoundResolver)(nil)

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

func (c *CompoundResolver) LookupComponentVersion(name string, version string) (*ComponentDescriptor, error) {
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
	return nil, errors.ErrNotFound(KIND_REFERENCE, common.NewNameVersion(name, version).String())
}

func (c *CompoundResolver) AddResolver(r ComponentVersionResolver) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if r != nil {
		c.resolvers = append(c.resolvers, r)
	}
}
