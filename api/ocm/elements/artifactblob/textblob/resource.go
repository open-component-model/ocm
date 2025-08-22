package textblob

import (
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/elements/artifactblob/datablob"
	"ocm.software/ocm/api/utils/mime"
)

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, blob string, opts ...Option) cpi.ArtifactAccess[M] {
	var eff Options
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyTo(&eff)
		}
	}

	if eff.MimeType == "" {
		eff.MimeType = mime.MIME_TEXT
	}
	return datablob.Access(ctx, meta, []byte(blob), &eff)
}

func ResourceAccess(ctx ocm.Context, meta *ocm.ResourceMeta, blob string, opts ...Option) cpi.ResourceAccess {
	return Access(ctx, meta, blob, opts...)
}

func SourceAccess(ctx ocm.Context, meta *ocm.SourceMeta, blob string, opts ...Option) cpi.SourceAccess {
	return Access(ctx, meta, blob, opts...)
}
