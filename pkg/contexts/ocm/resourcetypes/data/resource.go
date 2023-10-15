// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"github.com/open-component-model/ocm/pkg/common/accessio/blobaccess"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes/rpi"
	"github.com/open-component-model/ocm/pkg/generics"
	"github.com/open-component-model/ocm/pkg/mime"
)

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, blob []byte, opts ...Option) cpi.ArtifactAccess[M] {
	eff := rpi.EvalOptions[Options](opts...)
	if eff.MimeType == "" {
		eff.MimeType = mime.MIME_OCTET
	}
	accprov := cpi.NewAccessProviderForBlobAccessProvider(ctx, blobaccess.ProviderForData(eff.MimeType, blob), eff.Hint, eff.Global)
	// strange type cast is required by Go compiler, meta has the correct type.
	return cpi.NewArtifactAccessForProvider(generics.As[*M](meta), accprov)
}

func ResourceAccess(ctx ocm.Context, media string, meta *ocm.ResourceMeta, blob []byte, opts ...Option) cpi.ResourceAccess {
	return Access(ctx, meta, blob, opts...)
}

func SourceAccess(ctx ocm.Context, media string, meta *ocm.SourceMeta, blob []byte, opts ...Option) cpi.SourceAccess {
	return Access(ctx, meta, blob, opts...)
}
