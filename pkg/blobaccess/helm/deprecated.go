package helm

import (
	"github.com/open-component-model/ocm/pkg/blobaccess/bpi"
)

// BlobAccessForHelmChart returns a BlobAccess for the Helm chart with the given path.
// Deprecated: use BlobAccess.
func BlobAccessForHelmChart(path string, opts ...Option) (blob bpi.BlobAccess, name, version string, err error) {
	return BlobAccess(path, opts...)
}

// BlobAccessProviderForHelmChart returns a BlobAccessProvider for the Helm chart with the given name.
// Deprecated: use Provider.
func BlobAccessProviderForHelmChart(name string, opts ...Option) bpi.BlobAccessProvider {
	return Provider(name, opts...)
}
