// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cpi

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	cpi "github.com/open-component-model/ocm/pkg/contexts/ocm/internal"
)

////////////////////////////////////////////////////////////////////////////////

type BaseAccess struct {
	vers   ComponentVersionAccess
	access compdesc.AccessSpec
}

type baseAccess = BaseAccess

func NewBaseAccess(cv ComponentVersionAccess, acc compdesc.AccessSpec) *BaseAccess {
	return &BaseAccess{vers: cv, access: acc}
}

func (r *BaseAccess) GetOCMContext() Context {
	return r.vers.GetContext()
}

func (r *BaseAccess) ReferenceHint() string {
	if hp, ok := r.access.(cpi.HintProvider); ok {
		return hp.GetReferenceHint(r.vers)
	}
	return ""
}

func (r *BaseAccess) Access() (AccessSpec, error) {
	return r.vers.GetContext().AccessSpecForSpec(r.access)
}

func (r *BaseAccess) AccessMethod() (AccessMethod, error) {
	acc, err := r.vers.GetContext().AccessSpecForSpec(r.access)
	if err != nil {
		return nil, err
	}
	return acc.AccessMethod(r.vers)
}

func (r *BaseAccess) BlobAccess() (BlobAccess, error) {
	m, err := r.AccessMethod()
	if err != nil {
		return nil, err
	}
	return BlobAccessForAccessMethod(AccessMethodAsView(m))
}

////////////////////////////////////////////////////////////////////////////////

type resourceAccessImpl struct {
	*baseAccess
	meta ResourceMeta
}

var _ ResourceAccess = (*resourceAccessImpl)(nil)

func NewResourceAccess(componentVersion ComponentVersionAccess, accessSpec compdesc.AccessSpec, meta ResourceMeta) ResourceAccess {
	return &resourceAccessImpl{
		baseAccess: NewBaseAccess(componentVersion, accessSpec),
		meta:       meta,
	}
}

func (r *resourceAccessImpl) Meta() *ResourceMeta {
	return &r.meta
}

////////////////////////////////////////////////////////////////////////////////

type sourceAccessImpl struct {
	*baseAccess
	meta SourceMeta
}

var _ SourceAccess = (*sourceAccessImpl)(nil)

func NewSourceAccess(componentVersion ComponentVersionAccess, accessSpec compdesc.AccessSpec, meta SourceMeta) SourceAccess {
	return &sourceAccessImpl{
		baseAccess: NewBaseAccess(componentVersion, accessSpec),
		meta:       meta,
	}
}

func (r sourceAccessImpl) Meta() *SourceMeta {
	return &r.meta
}
