// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package text

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/elements/localartifacts/data"
)

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, blob string, opts ...Option) cpi.ArtifactAccess[M] {
	return data.Access(ctx, meta, []byte(blob), opts...)
}

func ResourceAccess(ctx ocm.Context, media string, meta *ocm.ResourceMeta, blob string, opts ...Option) cpi.ResourceAccess {
	return Access(ctx, meta, blob, opts...)
}

func SourceAccess(ctx ocm.Context, media string, meta *ocm.SourceMeta, blob string, opts ...Option) cpi.SourceAccess {
	return Access(ctx, meta, blob, opts...)
}
