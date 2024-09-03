package npm

import (
	"ocm.software/ocm/api/utils/blobaccess/bpi"
)

func DataAccess(repo string, pkg, version string, opts ...Option) (bpi.DataAccess, error) {
	return BlobAccess(repo, pkg, version, opts...)
}

func BlobAccess(repo string, pkg, version string, opts ...Option) (bpi.BlobAccess, error) {
	s, err := NewPackageSpec(repo, pkg, version, opts...)
	if err != nil {
		return nil, err
	}
	return s.GetBlobAccess()
}

func Provider(repo string, pkg, version string, opts ...Option) bpi.BlobAccessProvider {
	return bpi.BlobAccessProviderFunction(func() (bpi.BlobAccess, error) {
		b, err := BlobAccess(repo, pkg, version, opts...)
		return b, err
	})
}
