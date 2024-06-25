package github

import (
	"github.com/mandelsoft/goutils/optionutils"

	"github.com/open-component-model/ocm/api/ocm"
	"github.com/open-component-model/ocm/api/ocm/compdesc"
	"github.com/open-component-model/ocm/api/ocm/cpi"
	"github.com/open-component-model/ocm/api/ocm/elements/artifactaccess/genericaccess"
	access "github.com/open-component-model/ocm/api/ocm/extensions/accessmethods/s3"
	"github.com/open-component-model/ocm/api/ocm/extensions/resourcetypes"
	"github.com/open-component-model/ocm/api/utils/mime"
)

const TYPE = resourcetypes.BLOB

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, bucket, key string, opts ...Option) cpi.ArtifactAccess[M] {
	eff := optionutils.EvalOptions(opts...)
	if meta.GetType() == "" {
		meta.SetType(TYPE)
	}

	media := eff.MediaType
	if media == "" {
		media = mime.MIME_OCTET
	}
	spec := access.New(eff.Region, bucket, key, eff.Version, media)
	// is global access, must work, otherwise there is an error in the lib.
	return genericaccess.MustAccess(ctx, meta, spec)
}

func ResourceAccess(ctx ocm.Context, meta *cpi.ResourceMeta, bucket, key string, opts ...Option) cpi.ResourceAccess {
	return Access(ctx, meta, bucket, key, opts...)
}

func SourceAccess(ctx ocm.Context, meta *cpi.SourceMeta, bucket, key string, opts ...Option) cpi.SourceAccess {
	return Access(ctx, meta, bucket, key, opts...)
}
