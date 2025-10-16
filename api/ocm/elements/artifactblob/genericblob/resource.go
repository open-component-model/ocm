package genericblob

import (
	"github.com/mandelsoft/goutils/generics"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
)

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx cpi.Context, meta P, blob blobaccess.BlobAccessProvider, opts ...Option) cpi.ArtifactAccess[M] {
	var eff Options
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyTo(&eff)
		}
	}
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
