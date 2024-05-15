package genericblob

import (
	"github.com/mandelsoft/goutils/generics"

	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/optionutils"
)

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx cpi.Context, meta P, blob blobaccess.BlobAccessProvider, opts ...Option) cpi.ArtifactAccess[M] {
	eff := optionutils.EvalOptions(opts...)
	accprov := cpi.NewAccessProviderForBlobAccessProvider(ctx, blob, eff.Hint, eff.Global)
	// strange type cast is required by Go compiler, meta has the correct type.
	return cpi.NewArtifactAccessForProvider(generics.Cast[*M](meta), accprov)
}

func ResourceAccess(ctx cpi.Context, media string, meta *cpi.ResourceMeta, blob blobaccess.BlobAccessProvider, opts ...Option) cpi.ResourceAccess {
	return Access(ctx, meta, blob, opts...)
}

func SourceAccess(ctx cpi.Context, media string, meta *cpi.SourceMeta, blob blobaccess.BlobAccessProvider, opts ...Option) cpi.SourceAccess {
	return Access(ctx, meta, blob, opts...)
}
