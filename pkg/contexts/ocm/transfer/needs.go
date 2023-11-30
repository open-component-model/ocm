// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"github.com/go-test/deep"

	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
)

func needsResourceTransport(cv ocm.ComponentVersionAccess, s, t *compdesc.ComponentDescriptor, handler TransferHandler) bool {
	for _, r := range s.Resources {
		rt, err := t.GetResourceByIdentity(r.GetIdentity(s.Resources))
		if err != nil {
			return true
		}

		sa := cpi.NewResourceAccess(cv, r.Access, r.ResourceMeta)
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

		sa := cpi.NewSourceAccess(cv, r.Access, r.SourceMeta)

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
