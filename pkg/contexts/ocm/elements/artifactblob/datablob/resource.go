// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package datablob

import (
	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/generics"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/optionutils"
)

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, blob []byte, opts ...Option) cpi.ArtifactAccess[M] {
	eff := optionutils.EvalOptions(opts...)

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
	return cpi.NewArtifactAccessForProvider(generics.As[*M](meta), accprov)
}

func ResourceAccess(ctx ocm.Context, media string, meta *ocm.ResourceMeta, blob []byte, opts ...Option) cpi.ResourceAccess {
	return Access(ctx, meta, blob, opts...)
}

func SourceAccess(ctx ocm.Context, media string, meta *ocm.SourceMeta, blob []byte, opts ...Option) cpi.SourceAccess {
	return Access(ctx, meta, blob, opts...)
}
