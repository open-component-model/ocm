package git

import (
	"ocm.software/ocm/api/oci/extensions/repositories/git"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm/extensions/repositories/genericocireg"
	"ocm.software/ocm/api/utils/accessobj"
)

type Object = ctf.Object

const (
	ACC_CREATE   = accessobj.ACC_CREATE
	ACC_WRITABLE = accessobj.ACC_WRITABLE
	ACC_READONLY = accessobj.ACC_READONLY
)

func Open(ctx cpi.ContextProvider, acc accessobj.AccessMode, url string, opts Options) (cpi.Repository, error) {
	r, err := git.Open(cpi.FromProvider(ctx), acc, url, opts)
	if err != nil {
		return nil, err
	}
	return genericocireg.NewRepository(cpi.FromProvider(ctx), nil, r), nil
}

func Create(ctx cpi.ContextProvider, acc accessobj.AccessMode, url string, opts Options) (cpi.Repository, error) {
	r, err := git.Create(cpi.FromProvider(ctx), acc, url, opts)
	if err != nil {
		return nil, err
	}
	return genericocireg.NewRepository(cpi.FromProvider(ctx), nil, r), nil
}
