// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package access

import (
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

func NewResourceAccess(cv ocm.ComponentVersionAccess, rsc *compdesc.Resource) ocm.ResourceAccess {
	return &resourceAccess{
		baseAccess: baseAccess{
			cv:     cv,
			access: rsc.Access,
		},
		resource: rsc,
	}
}

func (r *resourceAccess) Meta() *compdesc.ResourceMeta {
	return &r.resource.ResourceMeta
}

type sourceAccess struct {
	baseAccess
	source *compdesc.Source
}

var _ ocm.SourceAccess = (*sourceAccess)(nil)

func NewSourceAccess(cv ocm.ComponentVersionAccess, sc *compdesc.Source) ocm.SourceAccess {
	return &sourceAccess{
		baseAccess: baseAccess{
			cv:     cv,
			access: sc.Access,
		},
		source: sc,
	}
}

func (r *sourceAccess) Meta() *compdesc.SourceMeta {
	return &r.source.SourceMeta
}
