package textblob

import (
	"github.com/mandelsoft/goutils/optionutils"

	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/elements/artifactblob/datablob"
	"github.com/open-component-model/ocm/pkg/mime"
)

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, blob string, opts ...Option) cpi.ArtifactAccess[M] {
	eff := optionutils.EvalOptions(opts...)
	if eff.MimeType == "" {
		eff.MimeType = mime.MIME_TEXT
	}
	return datablob.Access(ctx, meta, []byte(blob), eff)
}

func ResourceAccess(ctx ocm.Context, meta *ocm.ResourceMeta, blob string, opts ...Option) cpi.ResourceAccess {
	return Access(ctx, meta, blob, opts...)
}

func SourceAccess(ctx ocm.Context, meta *ocm.SourceMeta, blob string, opts ...Option) cpi.SourceAccess {
	return Access(ctx, meta, blob, opts...)
}
