package git

import (
	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
)

// //////////////////////////////////////////////////////////////////////////////

func Open(ctx cpi.ContextProvider, acc accessobj.AccessMode, url string, option ...accessio.Option) (Repository, error) {
	spec, err := NewRepositorySpec(acc, url, option...)
	if err != nil {
		return nil, err
	}
	return New(cpi.FromProvider(ctx), spec)
}

func Create(ctx cpi.ContextProvider, acc accessobj.AccessMode, url string, option ...accessio.Option) (Repository, error) {
	return Open(ctx, acc, url, option...)
}
