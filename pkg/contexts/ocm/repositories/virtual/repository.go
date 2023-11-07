// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package virtual

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi/repocpi"
	"github.com/open-component-model/ocm/pkg/refmgmt"
)

type _RepositoryImplBase = repocpi.RepositoryImplBase

type RepositoryImpl struct {
	_RepositoryImplBase
	access Access
	nonref cpi.Repository
}

var _ repocpi.RepositoryImpl = (*RepositoryImpl)(nil)

func NewRepository(ctx cpi.Context, acc Access) cpi.Repository {
	impl := &RepositoryImpl{
		_RepositoryImplBase: *repocpi.NewRepositoryImplBase(ctx.OCMContext()),
		access:              acc,
	}
	impl.nonref = repocpi.NewNoneRefRepositoryView(impl)
	r := repocpi.NewRepository(impl, "OCM repo[Simple]")
	return r
}

/*
func (r *RepositoryImpl) GetConsumerId(uctx ...credentials.UsageContext) credentials.ConsumerIdentity {
	return nil
}

func (r *RepositoryImpl) GetIdentityMatcher() string {
	return ""
}
*/

func (r *RepositoryImpl) Close() error {
	return r.access.Close()
}

func (r *RepositoryImpl) GetSpecification() cpi.RepositorySpec {
	if p, ok := r.access.(RepositorySpecProvider); ok {
		return p.GetSpecification()
	}
	return NewRepositorySpec(r.access)
}

func (r *RepositoryImpl) ComponentLister() cpi.ComponentLister {
	return r.access.ComponentLister()
}

func (r *RepositoryImpl) ExistsComponentVersion(name string, version string) (bool, error) {
	return r.access.ExistsComponentVersion(name, version)
}

func (r *RepositoryImpl) LookupComponent(name string) (cpi.ComponentAccess, error) {
	return newComponentAccess(r, name, true)
}

func (r *RepositoryImpl) LookupComponentVersion(name string, version string) (cpi.ComponentVersionAccess, error) {
	c, err := newComponentAccess(r, name, false)
	if err != nil {
		return nil, err
	}
	defer refmgmt.PropagateCloseTemporary(&err, c) // temporary component object not exposed.
	return c.LookupVersion(version)
}
