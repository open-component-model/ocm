package ocmaccess

import (
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/selectors/rscsel"
	"ocm.software/ocm/api/utils/blobaccess/ocm"
)

////////////////////////////////////////////////////////////////////////////////
// Component Version

func ByComponentVersion(cv cpi.ComponentVersionAccess) ocm.ComponentVersionProvider {
	return ocm.ByComponentVersion(cv)
}

func ByResolverAndName(resolver cpi.ComponentVersionResolver, comp, vers string) ocm.ComponentVersionProvider {
	return ocm.ByResolverAndName(resolver, comp, vers)
}

func ByRepositorySpecAndName(ctx cpi.ContextProvider, spec cpi.RepositorySpec, comp, vers string) ocm.ComponentVersionProvider {
	return ocm.ByRepositorySpecAndName(ctx, spec, comp, vers)
}

////////////////////////////////////////////////////////////////////////////////
// Resource

func ByResourceId(id metav1.Identity) ocm.ResourceProvider {
	return ocm.ByResourceId(id)
}

func ByResourcePath(id metav1.Identity, path ...metav1.Identity) ocm.ResourceProvider {
	return ocm.ByResourcePath(id, path...)
}

func ByResourceSelector(sel ...rscsel.Selector) ocm.ResourceProvider {
	return ocm.ByResourceSelector(sel...)
}
