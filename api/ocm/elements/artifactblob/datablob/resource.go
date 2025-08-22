package datablob

import (
	"github.com/mandelsoft/goutils/generics"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/mime"
)

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, blob []byte, opts ...Option) cpi.ArtifactAccess[M] {
	var eff Options
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyTo(&eff)
		}
	}
	media := eff.MimeType
	if media == "" {
		media = mime.MIME_OCTET
	}

	var blobprov blobaccess.BlobAccessProvider
	switch eff.Compression {
	case NONE:
		blobprov = blobaccess.ProviderForData(media, blob)
	case COMPRESSION:
		blob := blobaccess.ForData(media, blob)
		defer blob.Close()
		blob, _ = blobaccess.WithCompression(blob)
		blobprov = blobaccess.ProviderForBlobAccess(blob)
	case DECOMPRESSION:
		blob := blobaccess.ForData(media, blob)
		defer blob.Close()
		blob, _ = blobaccess.WithDecompression(blob)
		blobprov = blobaccess.ProviderForBlobAccess(blob)
	}

	accprov := cpi.NewAccessProviderForBlobAccessProvider(ctx, blobprov, eff.Hint, eff.Global)
	// strange type cast is required by Go compiler, meta has the correct type.
	return cpi.NewArtifactAccessForProvider(generics.Cast[*M](meta), accprov)
}

func ResourceAccess(ctx ocm.Context, media string, meta *ocm.ResourceMeta, blob []byte, opts ...Option) cpi.ResourceAccess {
	return Access(ctx, meta, blob, opts...)
}

func SourceAccess(ctx ocm.Context, media string, meta *ocm.SourceMeta, blob []byte, opts ...Option) cpi.SourceAccess {
	return Access(ctx, meta, blob, opts...)
}
