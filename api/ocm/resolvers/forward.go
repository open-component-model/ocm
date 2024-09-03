package resolvers

import (
	"ocm.software/ocm/api/ocm/internal"
)

type (
	ContextProvider                  = internal.ContextProvider
	RepositorySpec                   = internal.RepositorySpec
	ComponentVersionAccess           = internal.ComponentVersionAccess
	ComponentVersionResolver         = internal.ComponentVersionResolver
	ComponentResolver                = internal.ComponentResolver
	Repository                       = internal.Repository
	ResolvedComponentVersionProvider = internal.ResolvedComponentVersionProvider
	ResolvedComponentProvider        = internal.ResolvedComponentProvider
	ResolvedRepositoryProvider       = internal.ResolvedComponentProvider
)

const (
	KIND_COMPONENTVERSION = internal.KIND_COMPONENTVERSION
	KIND_COMPONENT        = internal.KIND_COMPONENT
	KIND_OCM_REFERENCE    = internal.KIND_OCM_REFERENCE
)
