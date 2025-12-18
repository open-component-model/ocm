package oci

import (
	"reflect"

	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"

	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/cpi"
	ocmcpi "ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/repositories/genericocireg/componentmapping"
)

// StorageContext is the context information passed for Blobhandler
// registered for context type oci.CONTEXT_TYPE.
type StorageContext struct {
	ocmcpi.DefaultStorageContext
	Repository cpi.Repository
	Namespace  cpi.NamespaceAccess
	Manifest   cpi.ManifestAccess
}

var _ ocmcpi.StorageContext = (*StorageContext)(nil)

func New(compname string, repo ocmcpi.Repository, impltyp string, ocirepo oci.Repository, namespace oci.NamespaceAccess, manifest oci.ManifestAccess) *StorageContext {
	return &StorageContext{
		DefaultStorageContext: *ocmcpi.NewDefaultStorageContext(
			repo,
			compname,
			ocmcpi.ImplementationRepositoryType{
				ContextType:    cpi.CONTEXT_TYPE,
				RepositoryType: impltyp,
			},
		),
		Repository: ocirepo,
		Namespace:  namespace,
		Manifest:   manifest,
	}
}

func (s *StorageContext) TargetComponentRepository() ocmcpi.Repository {
	return s.ComponentRepository
}

func (s *StorageContext) TargetComponentName() string {
	return s.ComponentName
}

func (s *StorageContext) AssureLayer(blob cpi.BlobAccess) error {
	return s.Manifest.Modify(func(manifest *artdesc.Manifest) error {
		return AssureLayerLocked(blob, manifest)
	})
}

// AssureLayerLocked assures that the blob is listed as a layer in the manifest.
func AssureLayerLocked(blob cpi.BlobAccess, manifest *artdesc.Manifest) error {
	d := artdesc.DefaultBlobDescriptor(blob)
	found := -1
	for i, l := range manifest.Layers {
		if reflect.DeepEqual(&manifest.Layers[i], d) {
			return nil
		}
		if l.Digest == blob.Digest() {
			found = i
		}
	}
	if found > 0 { // ignore layer 0 used for component descriptor
		manifest.Layers[found] = *d
	} else {
		if len(manifest.Layers) == 0 {
			// fake descriptor layer
			manifest.Layers = append(manifest.Layers, ociv1.Descriptor{MediaType: componentmapping.ComponentDescriptorConfigMimeType})
		}
		manifest.Layers = append(manifest.Layers, *d)
	}
	return nil
}
