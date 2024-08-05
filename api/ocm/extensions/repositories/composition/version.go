package composition

import (
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/utils/refmgmt"
)

func NewComponentVersion(ctx cpi.ContextProvider, name, vers string) cpi.ComponentVersionAccess {
	repo := NewRepository(ctx)
	defer repo.Close()
	if !refmgmt.Lazy(repo) {
		panic("wrong composition repo implementation")
	}
	c, err := repo.LookupComponent(name)
	if err != nil {
		panic("wrong composition repo implementation: " + err.Error())
	}
	defer c.Close()
	cv, err := c.NewVersion(vers)
	if err != nil {
		panic("wrong composition repo implementation: " + err.Error())
	}
	return cv
}
