package ocm

import (
	"context"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/core"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/runtime"

	"github.com/opencontainers/go-digest"
)

const KIND_COMPONENTVERSION = core.KIND_COMPONENTVERSION
const KIND_RESOURCE = "component resource"
const KIND_SOURCE = "component source"
const KIND_REFERENCE = "component reference"

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
type HintProvider = core.HintProvider
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
type ComponentReference = core.ComponentReference

type DigesterType = core.DigesterType
type BlobDigester = core.BlobDigester
type BlobDigesterRegistry = core.BlobDigesterRegistry
type DigestDescriptor = core.DigestDescriptor

type BlobHandlerRegistry = core.BlobHandlerRegistry
type BlobHandler = core.BlobHandler

func NewDigestDescriptor(digest digest.Digest, typ DigesterType) *DigestDescriptor {
	return core.NewDigestDescriptor(digest, typ)
}

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

func NewGenericAccessSpec(spec string) (AccessSpec, error) {
	return core.NewGenericAccessSpec(spec)
}

type AccessSpecRef = core.AccessSpecRef

func NewAccessSpecRef(spec cpi.AccessSpec) *AccessSpecRef {
	return core.NewAccessSpecRef(spec)
}

func NewRawAccessSpecRef(data []byte, unmarshaler runtime.Unmarshaler) (*AccessSpecRef, error) {
	return core.NewRawAccessSpecRef(data, unmarshaler)
}
