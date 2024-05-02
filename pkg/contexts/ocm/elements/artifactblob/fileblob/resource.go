package fileblob

import (
	"github.com/mandelsoft/goutils/generics"

	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/optionutils"
)

const TYPE = "blob"

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, media string, meta P, path string, opts ...Option) cpi.ArtifactAccess[M] {
	eff := optionutils.EvalOptions(opts...)

	if meta.GetType() == "" {
		meta.SetType(TYPE)
	}
	if media == "" {
		media = mime.MIME_OCTET
	}

	var blobprov blobaccess.BlobAccessProvider
	switch eff.Compression {
	case NONE:
		blobprov = blobaccess.ProviderForFile(media, path, eff.FileSystem)
	case COMPRESSION:
		blob := blobaccess.ForFile(media, path, eff.FileSystem)
		defer blob.Close()
		blob, _ = blobaccess.WithCompression(blob)
		blobprov = blobaccess.ProviderForBlobAccess(blob)
	case DECOMPRESSION:
		blob := blobaccess.ForFile(media, path, eff.FileSystem)
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
