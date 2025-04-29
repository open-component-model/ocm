package ocireg

import (
	"github.com/mandelsoft/goutils/general"

	"ocm.software/ocm/api/oci/extensions/repositories/ocireg"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/repositories/genericocireg"
)

// ComponentNameMapping describes the method that is used to map the "Component Name", "Component Version"-tuples
// to OCI Image References.
type ComponentNameMapping = genericocireg.ComponentNameMapping

const (
	Type   = ocireg.Type
	TypeV1 = ocireg.TypeV1

	OCIRegistryURLPathMapping ComponentNameMapping = "urlPath"
	OCIRegistryDigestMapping  ComponentNameMapping = "sha256-digest"
)

// ComponentRepositoryMeta describes config special for a mapping of
// a component repository to an oci registry.
type ComponentRepositoryMeta = genericocireg.ComponentRepositoryMeta

// RepositorySpec describes a component repository backed by a oci registry.
type RepositorySpec = genericocireg.RepositorySpec

// NewRepositorySpec creates a new RepositorySpec.
// If no ocm meta is given, the subPath part is extracted from the base URL.
// Otherwise, the given URL is used as OCI registry URL as it is.
func NewRepositorySpec(baseURL string, metas ...*ComponentRepositoryMeta) *RepositorySpec {
	return genericocireg.NewRepositorySpec(ocireg.NewRepositorySpec(baseURL), general.Optional(metas...))
}

func NewComponentRepositoryMeta(subPath string, mapping ...ComponentNameMapping) *ComponentRepositoryMeta {
	return genericocireg.NewComponentRepositoryMeta(subPath, general.OptionalDefaulted(OCIRegistryURLPathMapping, mapping...))
}

func NewRepository(ctx cpi.ContextProvider, baseURL string, metas ...*ComponentRepositoryMeta) (cpi.Repository, error) {
	spec := NewRepositorySpec(baseURL, metas...)
	return ctx.OCMContext().RepositoryForSpec(spec)
}
