// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ociartifact

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/elements/externalartifacts/generic"
	"github.com/open-component-model/ocm/pkg/optionutils"
)

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, access ocm.AccessSpec, opts ...Option) (cpi.ArtifactAccess[M], error) {
	eff := optionutils.EvalOptions(opts...)

	hint := eff.Hint
	if hint == "" {
		hint = ocm.ReferenceHint(access, &cpi.DummyComponentVersionAccess{ctx})
	}
	global := eff.Global
	if global == nil {
		global = ocm.GlobalAccess(access, ctx)
	}

	a, err := generic.Access(ctx, meta, access)
	if err != nil {
		return nil, err
	}
	return newAccessProvider[M](a, hint, global), nil
}

type accessProvider[M any] struct {
	cpi.ArtifactAccess[M]
	hint   string
	global cpi.AccessSpec
}

func newAccessProvider[M any](prov cpi.ArtifactAccess[M], hint string, global cpi.AccessSpec) cpi.ArtifactAccess[M] {
	return &accessProvider[M]{
		ArtifactAccess: prov,
		hint:           hint,
		global:         global,
	}
}

func (p *accessProvider[M]) AccessSpec() cpi.AccessSpec {
	return nil
}

func (p *accessProvider[M]) ReferenceHint() string {
	if p.hint != "" {
		return p.hint
	}
	return p.ArtifactAccess.ReferenceHint()
}

func (p *accessProvider[M]) GlobalAccess() cpi.AccessSpec {
	if p.global != nil {
		return p.global
	}
	return p.ArtifactAccess.GlobalAccess()
}

func ResourceAccess(ctx ocm.Context, meta *cpi.ResourceMeta, access cpi.AccessSpec, opts ...Option) (cpi.ResourceAccess, error) {
	return Access(ctx, meta, access, opts...)
}

func SourceAccess(ctx ocm.Context, meta *cpi.SourceMeta, access cpi.AccessSpec, opts ...Option) (cpi.SourceAccess, error) {
	return Access(ctx, meta, access, opts...)
}
