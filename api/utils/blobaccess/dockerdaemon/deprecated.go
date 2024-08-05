package dockerdaemon

import (
	"ocm.software/ocm/api/utils/blobaccess/bpi"
)

// BlobAccessProviderForImageFromDockerDaemon returns a BlobAccessProvider for the image with the given name.
// Deprecated: use Provider.
func BlobAccessProviderForImageFromDockerDaemon(name string, opts ...Option) bpi.BlobAccessProvider {
	return Provider(name, opts...)
}

// BlobAccessForImageFromDockerDaemon returns a BlobAccess for the image with the given name.
// Decrecated: use BlobAccess.
func BlobAccessForImageFromDockerDaemon(name string, opts ...Option) (bpi.BlobAccess, string, error) {
	return BlobAccess(name, opts...)
}
