// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package genericaccess

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/generics"
)

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, access ocm.AccessSpec) (cpi.ArtifactAccess[M], error) {
	prov, err := cpi.NewAccessProviderForExternalAccessSpec(ctx, access)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid external access method %q", access.GetKind())
	}
	return cpi.NewArtifactAccessForProvider(generics.As[*M](meta), prov), nil
}

func MustAccess[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, access ocm.AccessSpec) cpi.ArtifactAccess[M] {
	a, err := Access(ctx, meta, access)
	if err != nil {
		panic(err)
	}
	return a
}

func ResourceAccess(ctx ocm.Context, meta *ocm.ResourceMeta, access ocm.AccessSpec) (cpi.ResourceAccess, error) {
	return Access(ctx, meta, access)
}

func SourceAccess(ctx ocm.Context, meta *ocm.SourceMeta, access ocm.AccessSpec) (cpi.SourceAccess, error) {
	return Access(ctx, meta, access)
}
