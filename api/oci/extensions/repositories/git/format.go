package git

import (
	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/utils/accessobj"
)

// //////////////////////////////////////////////////////////////////////////////

func Open(ctxp cpi.ContextProvider, acc accessobj.AccessMode, url string, opts Options) (Repository, error) {
	ctx := cpi.FromProvider(ctxp)
	spec, err := NewRepositorySpecFromOptions(acc, url, opts)
	if err != nil {
		return nil, err
	}
	return New(ctx, spec, nil)
}

func Create(ctx cpi.ContextProvider, url string, opts Options) (Repository, error) {
	return Open(ctx, accessobj.ACC_CREATE|accessobj.ACC_WRITABLE, url, opts)
}
