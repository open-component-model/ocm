package context

import (
	"ocm.software/ocm/api/ocm/internal"
)

type (
	Context                          = internal.Context
	ContextProvider                  = internal.ContextProvider
	LocalContextProvider             = internal.LocalContextProvider
	ComponentVersionResolver         = internal.ComponentVersionResolver
	ComponentResolver                = internal.ComponentResolver
	ResolvedComponentProvider        = internal.ResolvedComponentProvider
	ResolvedComponentVersionProvider = internal.ResolvedComponentVersionProvider
	Repository                       = internal.Repository
	RepositorySpecHandlers           = internal.RepositorySpecHandlers
	RepositorySpecHandler            = internal.RepositorySpecHandler
	RepositoryProvider               = internal.ResolvedComponentProvider
	UniformRepositorySpec            = internal.UniformRepositorySpec
	ComponentLister                  = internal.ComponentLister
	ComponentAccess                  = internal.ComponentAccess
	ComponentVersionAccess           = internal.ComponentVersionAccess
	AccessSpec                       = internal.AccessSpec
	GenericAccessSpec                = internal.GenericAccessSpec
	HintProvider                     = internal.HintProvider
	AccessMethod                     = internal.AccessMethodImpl
	AccessType                       = internal.AccessType
	DataAccess                       = internal.DataAccess
	BlobAccess                       = internal.BlobAccess
	SourceAccess                     = internal.SourceAccess
	SourceMeta                       = internal.SourceMeta
	ResourceAccess                   = internal.ResourceAccess
	ResourceMeta                     = internal.ResourceMeta
	RepositorySpec                   = internal.RepositorySpec
	GenericRepositorySpec            = internal.GenericRepositorySpec
	IntermediateRepositorySpecAspect = internal.IntermediateRepositorySpecAspect
	RepositoryType                   = internal.RepositoryType
	RepositoryTypeScheme             = internal.RepositoryTypeScheme
	RepositoryDelegationRegistry     = internal.RepositoryDelegationRegistry
	AccessTypeScheme                 = internal.AccessTypeScheme
	ComponentReference               = internal.ComponentReference
)

type (
	DigesterType         = internal.DigesterType
	BlobDigester         = internal.BlobDigester
	BlobDigesterRegistry = internal.BlobDigesterRegistry
	DigestDescriptor     = internal.DigestDescriptor
	HasherProvider       = internal.HasherProvider
	Hasher               = internal.Hasher
)

type (
	BlobHandlerRegistry = internal.BlobHandlerRegistry
	BlobHandler         = internal.BlobHandler
)
