// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package helmblob

import (
	"github.com/open-component-model/ocm/pkg/blobaccess/helm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/generics"
	"github.com/open-component-model/ocm/pkg/optionutils"
)

const TYPE = resourcetypes.HELM_CHART

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, path string, opts ...Option) cpi.ArtifactAccess[M] {
	eff := optionutils.EvalOptions(append(opts, WithContext(ctx))...)
	if meta.GetType() == "" {
		meta.SetType(TYPE)
	}
	hint := eff.Hint
	blobprov := helm.BlobAccessProviderForHelmChart(path, &eff.Blob)
	accprov := cpi.NewAccessProviderForBlobAccessProvider(ctx, blobprov, hint, eff.Global)
	// strange type cast is required by Go compiler, meta has the correct type.
	return cpi.NewArtifactAccessForProvider(generics.As[*M](meta), accprov)
}

func ResourceAccess(ctx ocm.Context, meta *cpi.ResourceMeta, path string, opts ...Option) cpi.ResourceAccess {
	return Access(ctx, meta, path, opts...)
}

func SourceAccess(ctx ocm.Context, meta *cpi.SourceMeta, path string, opts ...Option) cpi.SourceAccess {
	return Access(ctx, meta, path, opts...)
}
