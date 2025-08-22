package fileblob

import (
	"github.com/mandelsoft/goutils/generics"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/blobaccess/file"
	"ocm.software/ocm/api/utils/mime"
)

const TYPE = "blob"

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, media string, meta P, path string, opts ...Option) cpi.ArtifactAccess[M] {
	var eff Options
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyTo(&eff)
		}
	}

	if meta.GetType() == "" {
		meta.SetType(TYPE)
	}
	if media == "" {
		media = mime.MIME_OCTET
	}

	var blobprov blobaccess.BlobAccessProvider
	switch eff.Compression {
	case NONE:
		blobprov = file.Provider(media, path, eff.FileSystem)
	case COMPRESSION:
		blob := file.BlobAccess(media, path, eff.FileSystem)
		defer blob.Close()
		blob, _ = blobaccess.WithCompression(blob)
		blobprov = blobaccess.ProviderForBlobAccess(blob)
	case DECOMPRESSION:
		blob := file.BlobAccess(media, path, eff.FileSystem)
		defer blob.Close()
		blob, _ = blobaccess.WithDecompression(blob)
		blobprov = blobaccess.ProviderForBlobAccess(blob)
	}

	accprov := cpi.NewAccessProviderForBlobAccessProvider(ctx, blobprov, eff.Hint, eff.Global)
	// strange type cast is required by Go compiler, meta has the correct type.
	return cpi.NewArtifactAccessForProvider(generics.Cast[*M](meta), accprov)
}

func ResourceAccess(ctx ocm.Context, media string, meta *ocm.ResourceMeta, path string, opts ...Option) cpi.ResourceAccess {
	return Access(ctx, media, meta, path, opts...)
}

func SourceAccess(ctx ocm.Context, media string, meta *ocm.SourceMeta, path string, opts ...Option) cpi.SourceAccess {
	return Access(ctx, media, meta, path, opts...)
}
