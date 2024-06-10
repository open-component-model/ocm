package dockermulti

import (
	"github.com/open-component-model/ocm/pkg/blobaccess/bpi"
)

// BlobAccessForMultiImageFromDockerDaemon returns a BlobAccess for the image with the given name.
// Deprecated: use BlobAccess.
func BlobAccessForMultiImageFromDockerDaemon(opts ...Option) (bpi.BlobAccess, error) {
	return BlobAccess(opts...)
}

// BlobAccessProviderForMultiImageFromDockerDaemon returns a BlobAccessProvider for the image with the given name.
// Deprecated: use Provider.
func BlobAccessProviderForMultiImageFromDockerDaemon(opts ...Option) bpi.BlobAccessProvider {
	return Provider(opts...)
}
