package wgetaccess

import (
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/elements/artifactaccess/genericaccess"
	access "ocm.software/ocm/api/ocm/extensions/accessmethods/wget"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
)

const TYPE = resourcetypes.BLOB

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, url string, opts ...Option) cpi.ArtifactAccess[M] {
	if meta.GetType() == "" {
		meta.SetType(TYPE)
	}

	spec := access.New(url, opts...)
	// is global access, must work, otherwise there is an error in the lib.
	return genericaccess.MustAccess(ctx, meta, spec)
}

func ResourceAccess(ctx ocm.Context, meta *cpi.ResourceMeta, url string, opts ...Option) cpi.ResourceAccess {
	return Access(ctx, meta, url, opts...)
}

func SourceAccess(ctx ocm.Context, meta *cpi.SourceMeta, url string, opts ...Option) cpi.SourceAccess {
	return Access(ctx, meta, url, opts...)
}
