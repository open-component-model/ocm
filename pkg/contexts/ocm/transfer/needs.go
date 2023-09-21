// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"github.com/go-test/deep"

	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
)

type baseAccess struct {
	cv     ocm.ComponentVersionAccess
	access compdesc.AccessSpec
}

func (r *baseAccess) ComponentVersion() ocm.ComponentVersionAccess {
	return r.cv
}

func (r *baseAccess) Access() (ocm.AccessSpec, error) {
	return r.cv.GetContext().AccessSpecForSpec(r.access)
}

func (r *baseAccess) AccessMethod() (ocm.AccessMethod, error) {
	acc, err := r.Access()
	if err != nil {
		return nil, err
	}
	return r.cv.AccessMethod(acc)
}

type resourceAccess struct {
	baseAccess
	resource *compdesc.Resource
}

var _ ocm.ResourceAccess = (*resourceAccess)(nil)

func (r *resourceAccess) Meta() *compdesc.ResourceMeta {
	return &r.resource.ResourceMeta
}

type sourceAccess struct {
	baseAccess
	source *compdesc.Source
}

var _ ocm.SourceAccess = (*sourceAccess)(nil)

func (r *sourceAccess) Meta() *compdesc.SourceMeta {
	return &r.source.SourceMeta
}

func needsResourceTransport(cv ocm.ComponentVersionAccess, s, t *compdesc.ComponentDescriptor, handler TransferHandler) bool {
	for _, r := range s.Resources {
		rt, err := t.GetResourceByIdentity(r.GetIdentity(s.Resources))
		if err != nil {
			return true
		}

		sa := &resourceAccess{
			baseAccess: baseAccess{
				cv:     cv,
				access: r.Access,
			},
			resource: &r,
		}

		sacc, err := sa.Access()
		if err != nil {
			return true
		}
		if needsTransport(cv.GetContext(), sa, &rt) {
			ok, err := handler.TransferResource(cv, sacc, sa)
			return ok || err != nil
		}
	}

	for _, r := range s.Sources {
		rt, err := t.GetSourceByIdentity(r.GetIdentity(s.Sources))
		if err != nil {
			return true
		}

		sa := &sourceAccess{
			baseAccess: baseAccess{
				cv:     cv,
				access: r.Access,
			},
			source: &r,
		}

		sacc, err := sa.Access()
		if err != nil {
			return true
		}
		if needsTransport(cv.GetContext(), sa, &rt) {
			ok, err := handler.TransferSource(cv, sacc, sa)
			return ok || err != nil
		}
	}
	return false
}

func needsTransport(ctx ocm.Context, s ocm.AccessProvider, t compdesc.AccessProvider) bool {
	sacc, err := s.Access()
	if err != nil {
		return true
	}
	tacc, err := ctx.AccessSpecForSpec(t.GetAccess())
	if err != nil {
		return true
	}

	if sacc.IsLocal(ctx) && tacc.IsLocal(ctx) {
		return false
	}
	return len(deep.Equal(sacc, tacc)) == 0
}
