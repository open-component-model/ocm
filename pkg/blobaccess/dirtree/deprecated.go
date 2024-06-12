package dirtree

import (
	"github.com/open-component-model/ocm/pkg/blobaccess/bpi"
)

// BlobAccessForDirTree returns a BlobAccess for the given directory tree.
// Deprecated: use BlobAccess.
func BlobAccessForDirTree(path string, opts ...Option) (_ bpi.BlobAccess, rerr error) {
	return BlobAccess(path, opts...)
}

// BlobAccessProviderForDirTree returns a BlobAccessProvider for the given directory tree.
// Deprecated: use Provider.
func BlobAccessProviderForDirTree(path string, opts ...Option) bpi.BlobAccessProvider {
	return Provider(path, opts...)
}
