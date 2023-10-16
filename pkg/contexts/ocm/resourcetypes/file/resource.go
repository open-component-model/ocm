// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package dirtree

import (
	"github.com/open-component-model/ocm/pkg/common/accessio/blobaccess"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/generics"
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

	blobprov := blobaccess.ProviderForFile(media, path, eff.FileSystem)
	accprov := cpi.NewAccessProviderForBlobAccessProvider(ctx, blobprov, eff.Hint, eff.Global)
	// strange type cast is required by Go compiler, meta has the correct type.
	return cpi.NewArtifactAccessForProvider(generics.As[*M](meta), accprov)
}

func ResourceAccess(ctx ocm.Context, media string, meta *ocm.ResourceMeta, path string, opts ...Option) cpi.ResourceAccess {
	return Access(ctx, media, meta, path, opts...)
}

func SourceAccess(ctx ocm.Context, media string, meta *ocm.SourceMeta, path string, opts ...Option) cpi.SourceAccess {
	return Access(ctx, media, meta, path, opts...)
}
