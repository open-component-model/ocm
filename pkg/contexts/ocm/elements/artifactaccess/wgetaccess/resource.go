package wgetaccess

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	access "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/wget"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/elements/artifactaccess/genericaccess"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/optionutils"
)

const TYPE = resourcetypes.BLOB

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, url, mimeType string, opts ...Option) cpi.ArtifactAccess[M] {
	eff := optionutils.EvalOptions(opts...)
	if meta.GetType() == "" {
		meta.SetType(TYPE)
	}

	media := eff.MimeType
	if media == "" {
		media = mime.MIME_OCTET
	}
	spec := access.New(url, mimeType)
	// is global access, must work, otherwise there is an error in the lib.
	return genericaccess.MustAccess(ctx, meta, spec)
}

func ResourceAccess(ctx ocm.Context, meta *cpi.ResourceMeta, url, mimeType string, opts ...Option) cpi.ResourceAccess {
	return Access(ctx, meta, url, mimeType, opts...)
}

func SourceAccess(ctx ocm.Context, meta *cpi.SourceMeta, bucket, key string, opts ...Option) cpi.SourceAccess {
	return Access(ctx, meta, bucket, key, opts...)
}
