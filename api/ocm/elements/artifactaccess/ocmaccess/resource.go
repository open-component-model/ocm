package ocmaccess

import (
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/elements/artifactaccess/genericaccess"
	access "ocm.software/ocm/api/ocm/extensions/accessmethods/ocm"
)

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, comp, vers string, repo cpi.RepositorySpec, id metav1.Identity, path ...metav1.Identity) (cpi.ArtifactAccess[M], error) {
	spec, err := access.New(comp, vers, repo, id, path...)
	if err != nil {
		return nil, err
	}
	// is global access, must work, otherwise there is an error in the lib.
	return genericaccess.MustAccess(ctx, meta, spec), nil
}

func ResourceAccess(ctx ocm.Context, meta *cpi.ResourceMeta, comp, vers string, repo cpi.RepositorySpec, id metav1.Identity, path ...metav1.Identity) (cpi.ResourceAccess, error) {
	return Access(ctx, meta, comp, vers, repo, id, path...)
}

func SourceAccess(ctx ocm.Context, meta *cpi.SourceMeta, comp, vers string, repo cpi.RepositorySpec, id metav1.Identity, path ...metav1.Identity) (cpi.SourceAccess, error) {
	return Access(ctx, meta, comp, vers, repo, id, path...)
}
