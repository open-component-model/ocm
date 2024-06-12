package wget

import (
	"github.com/open-component-model/ocm/pkg/blobaccess/bpi"
)

// DataAccessForWget returns a DataAccess for the given URL.
// Deprecated: use DataAccess.
func DataAccessForWget(url string, opts ...Option) (bpi.DataAccess, error) {
	return DataAccess(url, opts...)
}

// BlobAccessForWget returns a BlobAccess for the given URL.
// Deprecated: use BlobAccess.
func BlobAccessForWget(url string, opts ...Option) (_ bpi.BlobAccess, rerr error) {
	return BlobAccess(url, opts...)
}

// BlobAccessProviderForWget returns a BlobAccessProvider for the given URL.
// Deprecated: use Provider.
func BlobAccessProviderForWget(url string, opts ...Option) bpi.BlobAccessProvider {
	return Provider(url, opts...)
}
