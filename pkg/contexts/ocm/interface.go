package ocm

import (
	"context"

	core2 "github.com/open-component-model/ocm/pkg/contexts/ocm/core"

	"github.com/opencontainers/go-digest"
)

const KIND_COMPONENTVERSION = core2.KIND_COMPONENTVERSION

const CONTEXT_TYPE = core2.CONTEXT_TYPE

const CommonTransportFormat = core2.CommonTransportFormat

type Context = core2.Context
type Repository = core2.Repository
type RepositorySpecHandlers = core2.RepositorySpecHandlers
type RepositorySpecHandler = core2.RepositorySpecHandler
type UniformRepositorySpec = core2.UniformRepositorySpec
type ComponentLister = core2.ComponentLister
type ComponentAccess = core2.ComponentAccess
type ComponentVersionAccess = core2.ComponentVersionAccess
type AccessSpec = core2.AccessSpec
type AccessMethod = core2.AccessMethod
type AccessType = core2.AccessType
type DataAccess = core2.DataAccess
type BlobAccess = core2.BlobAccess
type SourceAccess = core2.SourceAccess
type SourceMeta = core2.SourceMeta
type ResourceAccess = core2.ResourceAccess
type ResourceMeta = core2.ResourceMeta
type RepositorySpec = core2.RepositorySpec
type RepositoryType = core2.RepositoryType
type RepositoryTypeScheme = core2.RepositoryTypeScheme
type AccessTypeScheme = core2.AccessTypeScheme
type BlobHandlerRegistry = core2.BlobHandlerRegistry
type ComponentReference = core2.ComponentReference

type DigesterType = core2.DigesterType
type BlobDigester = core2.BlobDigester
type BlobDigesterRegistry = core2.BlobDigesterRegistry
type DigestDescriptor = core2.DigestDescriptor

func NewDigestDescriptor(digest digest.Digest, typ DigesterType) *DigestDescriptor {
	return core2.NewDigestDescriptor(digest, typ)
}

// DefaultContext is the default context initialized by init functions
func DefaultContext() core2.Context {
	return core2.DefaultContext
}

func DefaultBlobHandlers() core2.BlobHandlerRegistry {
	return core2.DefaultBlobHandlerRegistry
}

// ForContext returns the Context to use for context.Context.
// This is eiter an explicit context or the default context.
func ForContext(ctx context.Context) Context {
	return core2.ForContext(ctx)
}

func NewGenericAccessSpec(spec string) (AccessSpec, error) {
	return core2.NewGenericAccessSpec(spec)
}
