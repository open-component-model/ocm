// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package generic

import (
	"github.com/open-component-model/ocm/pkg/common/accessio/blobaccess"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes/rpi"
	"github.com/open-component-model/ocm/pkg/generics"
)

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, blob blobaccess.BlobAccess, opts ...Option) cpi.ArtifactAccess[M] {
	eff := rpi.EvalOptions[Options](opts...)
	accprov := cpi.NewAccessProviderForBlobAccessProvider(ctx, blobaccess.ProviderForBlobAccess(blob), eff.Hint, eff.Global)
	// strange type cast is required by Go compiler, meta has the correct type.
	return cpi.NewArtifactAccessForProvider(generics.As[*M](meta), accprov)
}

func ResourceAccess(ctx ocm.Context, media string, meta *ocm.ResourceMeta, blob blobaccess.BlobAccess, opts ...Option) cpi.ResourceAccess {
	return Access(ctx, meta, blob, opts...)
}

func SourceAccess(ctx ocm.Context, media string, meta *ocm.SourceMeta, blob blobaccess.BlobAccess, opts ...Option) cpi.SourceAccess {
	return Access(ctx, meta, blob, opts...)
}
