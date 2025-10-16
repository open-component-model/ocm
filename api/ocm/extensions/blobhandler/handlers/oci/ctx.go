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
	return AssureLayer(s.Manifest.GetDescriptor(), blob)
}

func AssureLayer(desc *artdesc.Manifest, blob cpi.BlobAccess) error {
	d := artdesc.DefaultBlobDescriptor(blob)

	found := -1
	for i, l := range desc.Layers {
		if reflect.DeepEqual(&desc.Layers[i], d) {
			return nil
		}
		if l.Digest == blob.Digest() {
			found = i
		}
	}
	if found > 0 { // ignore layer 0 used for component descriptor
		desc.Layers[found] = *d
	} else {
		if len(desc.Layers) == 0 {
			// fake descriptor layer
			desc.Layers = append(desc.Layers, ociv1.Descriptor{MediaType: componentmapping.ComponentDescriptorConfigMimeType})
		}
		desc.Layers = append(desc.Layers, *d)
	}
	return nil
}
