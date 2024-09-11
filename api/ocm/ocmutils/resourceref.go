package ocmutils

import (
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	ocm "ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/resourcerefs"
)

// Deprectated: use resourcerefs.ResolveReferencePath.
func ResolveReferencePath(cv ocm.ComponentVersionAccess, path []metav1.Identity, resolver ocm.ComponentVersionResolver) (ocm.ComponentVersionAccess, error) {
	return resourcerefs.ResolveReferencePath(cv, path, resolver)
}

// Deprecated: use resourcerefs.MatchResourceReference.
func MatchResourceReference(cv ocm.ComponentVersionAccess, typ string, ref metav1.ResourceReference, resolver ocm.ComponentVersionResolver) (ocm.ResourceAccess, ocm.ComponentVersionAccess, error) {
	return resourcerefs.MatchResourceReference(cv, typ, ref, resolver)
}

// Deprecated: use resourcerefs.ResolveResourceReference.
func ResolveResourceReference(cv ocm.ComponentVersionAccess, ref metav1.ResourceReference, resolver ocm.ComponentVersionResolver) (ocm.ResourceAccess, ocm.ComponentVersionAccess, error) {
	return resourcerefs.ResolveResourceReference(cv, ref, resolver)
}
