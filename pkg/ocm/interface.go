package ocm

import (
	"context"

	"github.com/gardener/ocm/pkg/oci"
	_ "github.com/gardener/ocm/pkg/oci/repositories"
	_ "github.com/gardener/ocm/pkg/ocm/accessmethods"
	_ "github.com/gardener/ocm/pkg/ocm/compdesc/versions"
	_ "github.com/gardener/ocm/pkg/ocm/repositories"
	"github.com/gardener/ocm/pkg/ocm/repositories/genericocireg"

	"github.com/gardener/ocm/pkg/ocm/core"
)

type Context = core.Context
type ComponentRepository = core.Repository
type ComponentAccess = core.ComponentAccess
type AccessSpec = core.AccessSpec
type AccessMethod = core.AccessMethod
type AccessType = core.AccessType
type DataAccess = core.DataAccess
type ResourceAccess = core.ResourceAccess
type RepositorySpec = core.RepositorySpec
type RepositoryType = core.RepositoryType

func NewDefaultContext(ctx context.Context) Context {
	repoScheme := core.NewRepositoryTypeScheme(genericocireg.NewOCIRepositoryBackendType(oci.NewDefaultContext(ctx)))
	repoScheme.AddKnownTypes(core.DefaultRepositoryTypeScheme)
	return core.NewDefaultContext(ctx, oci.NewDefaultContext(ctx), core.DefaultAccessTypeScheme, repoScheme)
}
