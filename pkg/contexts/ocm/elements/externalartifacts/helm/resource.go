// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package helm

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	access "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/helm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/elements/externalartifacts/generic"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
)

const TYPE = resourcetypes.HELM_CHART

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, chart string, repourl string) cpi.ArtifactAccess[M] {
	if meta.GetType() == "" {
		meta.SetType(TYPE)
	}

	spec := access.New(chart, repourl)
	// is global access, must work, otherwise there is an error in the lib.
	return generic.MustAccess(ctx, meta, spec)
}

func ResourceAccess(ctx ocm.Context, meta *cpi.ResourceMeta, chart string, repourl string) cpi.ResourceAccess {
	return Access(ctx, meta, chart, repourl)
}

func SourceAccess(ctx ocm.Context, meta *cpi.SourceMeta, chart string, repourl string) cpi.SourceAccess {
	return Access(ctx, meta, chart, repourl)
}
