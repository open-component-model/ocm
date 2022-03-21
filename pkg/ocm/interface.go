package ocm

import (
	"context"

	_ "github.com/gardener/ocm/pkg/oci/repositories"
	_ "github.com/gardener/ocm/pkg/ocm/accessmethods"
	_ "github.com/gardener/ocm/pkg/ocm/blobhandler"
	_ "github.com/gardener/ocm/pkg/ocm/compdesc/versions"
	_ "github.com/gardener/ocm/pkg/ocm/repositories"

	"github.com/gardener/ocm/pkg/ocm/core"
)

const KIND_COMPONENTVERSION = core.KIND_COMPONENTVERSION

const CONTEXT_TYPE = core.CONTEXT_TYPE

const CommonTransportFormat = core.CommonTransportFormat

type Context = core.Context
type Repository = core.Repository
type RepositorySpecHandlers = core.RepositorySpecHandlers
type RepositorySpecHandler = core.RepositorySpecHandler
type UniformRepositorySpec = core.UniformRepositorySpec
type ComponentLister = core.ComponentLister
type ComponentAccess = core.ComponentAccess
type ComponentVersionAccess = core.ComponentVersionAccess
type AccessSpec = core.AccessSpec
type AccessMethod = core.AccessMethod
type AccessType = core.AccessType
type DataAccess = core.DataAccess
type BlobAccess = core.BlobAccess
type SourceAccess = core.SourceAccess
type SourceMeta = core.SourceMeta
type ResourceAccess = core.ResourceAccess
type ResourceMeta = core.ResourceMeta
type RepositorySpec = core.RepositorySpec
type RepositoryType = core.RepositoryType
type RepositoryTypeScheme = core.RepositoryTypeScheme
type AccessTypeScheme = core.AccessTypeScheme
type BlobHandlerRegistry = core.BlobHandlerRegistry

// DefaultContext is the default context initialized by init functions
func DefaultContext() core.Context {
	return core.DefaultContext
}

func DefaultBlobHandlers() core.BlobHandlerRegistry {
	return core.DefaultBlobHandlerRegistry
}

// ForContext returns the Context to use for context.Context.
// This is eiter an explicit context or the default context.
func ForContext(ctx context.Context) Context {
	return core.ForContext(ctx)
}
