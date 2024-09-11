package ociartifact

import (
	"ocm.software/ocm/api/utils/blobaccess/bpi"
)

// BlobAccessForOCIArtifact returns a BlobAccess for the OCI artifact with the given refname.
// Deprecated: use BlobAccess.
func BlobAccessForOCIArtifact(refname string, opts ...Option) (bpi.BlobAccess, string, error) {
	return BlobAccess(refname, opts...)
}

// BlobAccessProviderForOCIArtifact returns a BlobAccessProvider for the OCI artifact with the given name.
// Deprecated: use Provider.
func BlobAccessProviderForOCIArtifact(name string, opts ...Option) bpi.BlobAccessProvider {
	return Provider(name, opts...)
}
