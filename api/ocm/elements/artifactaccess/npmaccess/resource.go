package npmaccess

import (
	"github.com/open-component-model/ocm/api/ocm"
	"github.com/open-component-model/ocm/api/ocm/compdesc"
	"github.com/open-component-model/ocm/api/ocm/cpi"
	"github.com/open-component-model/ocm/api/ocm/elements/artifactaccess/genericaccess"
	access "github.com/open-component-model/ocm/api/ocm/extensions/accessmethods/npm"
	"github.com/open-component-model/ocm/api/ocm/extensions/resourcetypes"
)

const TYPE = resourcetypes.NPM_PACKAGE

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, registry, pkg, version string) cpi.ArtifactAccess[M] {
	if meta.GetType() == "" {
		meta.SetType(TYPE)
	}

	spec := access.New(registry, pkg, version)
	// is global access, must work, otherwise there is an error in the lib.
	return genericaccess.MustAccess(ctx, meta, spec)
}

func ResourceAccess(ctx ocm.Context, meta *cpi.ResourceMeta, registry, pkg, version string) cpi.ResourceAccess {
	return Access(ctx, meta, registry, pkg, version)
}

func SourceAccess(ctx ocm.Context, meta *cpi.SourceMeta, registry, pkg, version string) cpi.SourceAccess {
	return Access(ctx, meta, registry, pkg, version)
}
