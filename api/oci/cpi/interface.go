package cpi

// This is the Context Provider Interface for credential providers

import (
	"github.com/opencontainers/go-digest"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/oci/internal"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
)

const CONTEXT_TYPE = internal.CONTEXT_TYPE

const CommonTransportFormat = internal.CommonTransportFormat

type (
	Context                          = internal.Context
	ContextProvider                  = internal.ContextProvider
	Repository                       = internal.Repository
	RepositorySpecHandlers           = internal.RepositorySpecHandlers
	RepositorySpecHandler            = internal.RepositorySpecHandler
	UniformRepositorySpec            = internal.UniformRepositorySpec
	RepositoryType                   = internal.RepositoryType
	RepositoryTypeProvider           = internal.RepositoryTypeProvider
	RepositoryTypeScheme             = internal.RepositoryTypeScheme
	RepositorySpec                   = internal.RepositorySpec
	IntermediateRepositorySpecAspect = internal.IntermediateRepositorySpecAspect
	GenericRepositorySpec            = internal.GenericRepositorySpec
	ArtifactAccess                   = internal.ArtifactAccess
	Artifact                         = internal.Artifact
	ArtifactSource                   = internal.ArtifactSource
	ArtifactSink                     = internal.ArtifactSink
	BlobSource                       = internal.BlobSource
	BlobSink                         = internal.BlobSink
	NamespaceLister                  = internal.NamespaceLister
	NamespaceAccess                  = internal.NamespaceAccess
	ManifestAccess                   = internal.ManifestAccess
	IndexAccess                      = internal.IndexAccess
	BlobAccess                       = internal.BlobAccess
	DataAccess                       = internal.DataAccess
	RepositorySource                 = internal.RepositorySource
	ConsumerIdentityProvider         = internal.ConsumerIdentityProvider
)

type Descriptor = ociv1.Descriptor

func DefaultContext() Context {
	return internal.DefaultContext
}

func New(m ...datacontext.BuilderMode) Context {
	return internal.Builder{}.New(m...)
}

func FromProvider(p ContextProvider) Context {
	return internal.FromProvider(p)
}

func RegisterRepositorySpecHandler(handler RepositorySpecHandler, types ...string) {
	internal.RegisterRepositorySpecHandler(handler, types...)
}

func ToGenericRepositorySpec(spec RepositorySpec) (*GenericRepositorySpec, error) {
	return internal.ToGenericRepositorySpec(spec)
}

func UniformRepositorySpecForHostURL(typ string, host string) *UniformRepositorySpec {
	return internal.UniformRepositorySpecForHostURL(typ, host)
}

const (
	KIND_OCIARTIFACT = internal.KIND_OCIARTIFACT
	KIND_MEDIATYPE   = blobaccess.KIND_MEDIATYPE
	KIND_BLOB        = blobaccess.KIND_BLOB
)

func ErrUnknownArtifact(name, version string) error {
	return internal.ErrUnknownArtifact(name, version)
}

func ErrBlobNotFound(digest digest.Digest) error {
	return blobaccess.ErrBlobNotFound(digest)
}

func IsErrBlobNotFound(err error) bool {
	return blobaccess.IsErrBlobNotFound(err)
}

// provide context interface for other files to avoid diffs in imports.
var (
	newStrictRepositoryTypeScheme = internal.NewStrictRepositoryTypeScheme
	defaultRepositoryTypeScheme   = internal.DefaultRepositoryTypeScheme
)
