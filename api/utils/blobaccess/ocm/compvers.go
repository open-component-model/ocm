package ocm

import (
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/utils/refmgmt"
)

// ComponentVersionProvider is a factory for component versions.
// Every call provides a separately closeable component versuion access.
// An implementation should not hold private views of objects,
// The life cycle of all those objects should be left to the
// creator of ComponentVersionProvider implementation. Therefore, it does not
// have a close method.
type ComponentVersionProvider interface {
	GetComponentVersionAccess() (cpi.ComponentVersionAccess, error)
}

////////////////////////////////////////////////////////////////////////////////

type bycv struct {
	cv cpi.ComponentVersionAccess
}

var _ ComponentVersionProvider = (*bycv)(nil)

func ByComponentVersion(cv cpi.ComponentVersionAccess) ComponentVersionProvider {
	return &bycv{cv}
}

func (c *bycv) GetComponentVersionAccess() (cpi.ComponentVersionAccess, error) {
	return c.cv.Dup()
}

////////////////////////////////////////////////////////////////////////////////

type byresolver struct {
	resolver cpi.ComponentVersionResolver
	comp     string
	vers     string
}

var _ ComponentVersionProvider = (*byresolver)(nil)

func ByResolverAndName(resolver cpi.ComponentVersionResolver, comp, vers string) ComponentVersionProvider {
	return &byresolver{resolver, comp, vers}
}

func (c *byresolver) GetComponentVersionAccess() (cpi.ComponentVersionAccess, error) {
	return c.resolver.LookupComponentVersion(c.comp, c.vers)
}

////////////////////////////////////////////////////////////////////////////////

type byrepospec struct {
	ctx  cpi.Context
	spec cpi.RepositorySpec
	comp string
	vers string
}

var _ ComponentVersionProvider = (*byrepospec)(nil)

func ByRepositorySpecAndName(ctx cpi.ContextProvider, spec cpi.RepositorySpec, comp, vers string) ComponentVersionProvider {
	if ctx == nil {
		ctx = cpi.DefaultContext()
	}
	return &byrepospec{ctx.OCMContext(), spec, comp, vers}
}

func (c *byrepospec) GetComponentVersionAccess() (cpi.ComponentVersionAccess, error) {
	repo, err := refmgmt.ToLazy(c.ctx.RepositoryForSpec(c.spec))
	if err != nil {
		return nil, err
	}
	defer repo.Close()
	return repo.LookupComponentVersion(c.comp, c.vers)
}
