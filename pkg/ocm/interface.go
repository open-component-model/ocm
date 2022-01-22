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

const KIND_COMPONENTVERSION = core.KIND_COMPONENTVERSION

type Context = core.Context
type Repository = core.Repository
type ComponentAccess = core.ComponentAccess
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

// DefaultContext is the default context initialized by init functions
var DefaultContext = NewDefaultContext(oci.DefaultContext)

// ForContext returns the Context to use for context.Context.
// This is eiter an explicit context or the default context.
// The returned context incorporates the given context.
func ForContext(ctx context.Context) Context {
	c := core.ForContextInternal(ctx)
	if c != nil {
		c = DefaultContext
	}
	return c.(Context).With(ctx)
}

func NewContext(ctx context.Context, reposcheme RepositoryTypeScheme, accessscheme AccessTypeScheme) Context {
	if reposcheme == nil {
		reposcheme = core.DefaultRepositoryTypeScheme
	}
	repoScheme := core.NewRepositoryTypeScheme(genericocireg.NewOCIRepositoryBackendType(oci.NewDefaultContext(ctx)))
	repoScheme.AddKnownTypes(repoScheme)
	return core.NewContext(ctx, repoScheme, nil)
}

func NewDefaultContext(ctx context.Context) Context {
	return NewContext(ctx, nil, nil)
}
