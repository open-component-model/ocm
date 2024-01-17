package wgetblob

import (
	"github.com/open-component-model/ocm/pkg/blobaccess/wget"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/generics"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/optionutils"
)

const TYPE = resourcetypes.BLOB

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, url string, opts ...Option) cpi.ArtifactAccess[M] {
	eff := optionutils.EvalOptions(append([]Option{WithCredentialContext(ctx), WithLoggingContext(ctx)}, opts...)...)

	if meta.GetType() == "" {
		meta.SetType(TYPE)
	}
	if eff.Blob.MimeType == "" {
		eff.Blob.MimeType = mime.MIME_OCTET
	}
	blobprov := wget.BlobAccessProviderForWget(url, &eff.Blob)
	accprov := cpi.NewAccessProviderForBlobAccessProvider(ctx, blobprov, eff.Hint, eff.Global)
	// strange type cast is required by Go compiler, meta has the correct type.
	return cpi.NewArtifactAccessForProvider(generics.As[*M](meta), accprov)
}

func ResourceAccess(ctx ocm.Context, meta *ocm.ResourceMeta, url string, opts ...Option) cpi.ResourceAccess {
	return Access(ctx, meta, url, opts...)
}

func SourceAccess(ctx ocm.Context, meta *ocm.SourceMeta, url string, opts ...Option) cpi.SourceAccess {
	return Access(ctx, meta, url, opts...)
}
