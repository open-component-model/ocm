package ocm

import (
	"golang.org/x/exp/slices"

	"ocm.software/ocm/api/ocm/internal"
	"ocm.software/ocm/api/ocm/resolvers"
)

// Deprecated: use resolvers.DedicatedResolver.
type DedicatedResolver = resolvers.DedicatedResolver

// Deprecated: use resolvers.NewDedicatedResolver.
func NewDedicatedResolver(cv ...ComponentVersionAccess) ComponentVersionResolver {
	return resolvers.DedicatedResolver(slices.Clone(cv))
}

// Deprecated: use resolvers.CompoundResolver.
type CompoundResolver = resolvers.CompoundResolver

// Deprecated: use resolvers.NewCompoundResolver.
func NewCompoundResolver(res ...ComponentVersionResolver) ComponentVersionResolver {
	return resolvers.NewCompoundResolver(res...)
}

// Deprecated: use resolvers.MatchingResolver.
type MatchingResolver = resolvers.MatchingResolver

// Deprecated: use resolvers.NewMatchingResolver.
func NewMatchingResolver(ctx ContextProvider) resolvers.MatchingResolver {
	return internal.NewMatchingResolver(ctx.OCMContext())
}

// Deprecated: use resolvers.ResolverRule.
type ResolverRule = internal.ResolverRule
