package helmaccess

import (
	"github.com/open-component-model/ocm/api/ocm"
	"github.com/open-component-model/ocm/api/ocm/compdesc"
	"github.com/open-component-model/ocm/api/ocm/cpi"
	"github.com/open-component-model/ocm/api/ocm/elements/artifactaccess/genericaccess"
	access "github.com/open-component-model/ocm/api/ocm/extensions/accessmethods/helm"
	"github.com/open-component-model/ocm/api/ocm/extensions/resourcetypes"
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
