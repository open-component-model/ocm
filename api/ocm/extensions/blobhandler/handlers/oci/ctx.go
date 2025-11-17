package oci

import (
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/cpi"
	ocmcpi "ocm.software/ocm/api/ocm/cpi"
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
	return s.Manifest.AssureLayer(blob)
}
