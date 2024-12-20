package oras

import (
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// defaultManifestMediaTypes contains the default set of manifests media types.
var defaultManifestMediaTypes = []string{
	"application/vnd.docker.distribution.manifest.v2+json",
	"application/vnd.docker.distribution.manifest.list.v2+json",
	"application/vnd.oci.artifact.manifest.v1+json",
	ocispec.MediaTypeImageManifest,
	ocispec.MediaTypeImageIndex,
}

// isManifest determines if the given descriptor points to a manifest.
func isManifest(manifestMediaTypes []string, desc ocispec.Descriptor) bool {
	if len(manifestMediaTypes) == 0 {
		manifestMediaTypes = defaultManifestMediaTypes
	}
	for _, mediaType := range manifestMediaTypes {
		if desc.MediaType == mediaType {
			return true
		}
	}
	return false
}
