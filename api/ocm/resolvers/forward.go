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
	ResolverRule                     = internal.ResolverRule
)

const (
	KIND_COMPONENTVERSION = internal.KIND_COMPONENTVERSION
	KIND_COMPONENT        = internal.KIND_COMPONENT
	KIND_OCM_REFERENCE    = internal.KIND_OCM_REFERENCE
)

func NewResolverRule(prefix string, spec RepositorySpec, prio ...int) ResolverRule {
	return internal.NewResolverRule(prefix, spec, prio...)
}
