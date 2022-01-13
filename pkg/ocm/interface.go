package ocm

import (
	_ "github.com/gardener/ocm/pkg/ocm/accessmethods"
	_ "github.com/gardener/ocm/pkg/ocm/compdesc/versions"
	_ "github.com/gardener/ocm/pkg/ocm/registries"

	"github.com/gardener/ocm/pkg/ocm/core"
)

type ComponentRepository = core.Repository
type ComponentAccess = core.ComponentAccess
type AccessSpec = core.AccessSpec
type AccessMethod = core.AccessMethod
type AccessType = core.AccessType
type DataAccess = core.DataAccess
type ResourceAccess = core.ResourceAccess
type RepositorySpec = core.RepositorySpec
type RepositoryType = core.RepositoryType
