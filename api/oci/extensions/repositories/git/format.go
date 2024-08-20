package git

import (
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
)

type Object = Repository

// //////////////////////////////////////////////////////////////////////////////

func Open(ctx cpi.ContextProvider, acc accessobj.AccessMode, url string) (Object, error) {
	return New(cpi.FromProvider(ctx), &RepositorySpec{
		URL:        url,
		AccessMode: acc,
	})
}

func Create(ctx cpi.ContextProvider, acc accessobj.AccessMode, url string, mode vfs.FileMode, option ...accessio.Option) (Object, error) {
	spec, err := NewRepositorySpec(acc, url, mode, option...)
	if err != nil {
		return nil, err
	}

	return New(cpi.FromProvider(ctx), spec)
}
