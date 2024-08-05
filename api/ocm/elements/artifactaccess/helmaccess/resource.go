package helmaccess

import (
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/elements/artifactaccess/genericaccess"
	access "ocm.software/ocm/api/ocm/extensions/accessmethods/helm"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
)

const TYPE = resourcetypes.HELM_CHART

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, chart string, repourl string) cpi.ArtifactAccess[M] {
	if meta.GetType() == "" {
		meta.SetType(TYPE)
	}

	spec := access.New(chart, repourl)
	// is global access, must work, otherwise there is an error in the lib.
	return genericaccess.MustAccess(ctx, meta, spec)
}

func ResourceAccess(ctx ocm.Context, meta *cpi.ResourceMeta, chart string, repourl string) cpi.ResourceAccess {
	return Access(ctx, meta, chart, repourl)
}

func SourceAccess(ctx ocm.Context, meta *cpi.SourceMeta, chart string, repourl string) cpi.SourceAccess {
	return Access(ctx, meta, chart, repourl)
}
