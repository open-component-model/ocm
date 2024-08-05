package npmaccess

import (
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/elements/artifactaccess/genericaccess"
	access "ocm.software/ocm/api/ocm/extensions/accessmethods/npm"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
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
