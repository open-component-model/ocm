package git

import (
	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/utils/accessobj"
)

// //////////////////////////////////////////////////////////////////////////////

func Open(ctx cpi.ContextProvider, acc accessobj.AccessMode, url string, opts Options) (Repository, error) {
	spec, err := NewRepositorySpecFromOptions(acc, url, opts)
	if err != nil {
		return nil, err
	}
	return New(cpi.FromProvider(ctx), spec, nil)
}

func Create(ctx cpi.ContextProvider, acc accessobj.AccessMode, url string, opts Options) (Repository, error) {
	return Open(ctx, acc, url, opts)
}
