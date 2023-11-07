// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package repocpi

import (
	"fmt"
	"io"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	cpi2 "github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/internal"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/refmgmt"
	"github.com/open-component-model/ocm/pkg/refmgmt/resource"
	"github.com/open-component-model/ocm/pkg/utils"
)

// View objects are the user facing generic implementations of the context interfaces.
// They are responsible to handle the reference counting and use
// shared implementations objects for th concrete type-specific implementations.
// Additionally, they are used to implement interface functionality which is
// common to all implementations and NOT dependent on the backend system technology.

var (
	ErrClosed      = resource.ErrClosed
	ErrTempVersion = fmt.Errorf("temporary component version cannot be updated")
)

////////////////////////////////////////////////////////////////////////////////

type _RepositoryView interface {
	resource.ResourceViewInt[cpi2.Repository] // here you have to redeclare
}

type RepositoryViewManager = resource.ViewManager[cpi2.Repository] // here you have to use an alias

type RepositoryImpl interface {
	resource.ResourceImplementation[cpi2.Repository]
	internal.RepositoryImpl
}

type _RepositoryImplBase = resource.ResourceImplBase[cpi2.Repository]

type RepositoryImplBase struct {
	_RepositoryImplBase
	ctx cpi2.Context
}

func (b *RepositoryImplBase) GetContext() cpi2.Context {
	return b.ctx
}

func NewRepositoryImplBase(ctx cpi2.Context, closer ...io.Closer) *RepositoryImplBase {
	base, _ := resource.NewResourceImplBase[cpi2.Repository, io.Closer](nil, closer...)
	return &RepositoryImplBase{
		_RepositoryImplBase: *base,
		ctx:                 ctx,
	}
}

type repositoryView struct {
	_RepositoryView
	impl RepositoryImpl
}

var (
	_ cpi2.Repository                      = (*repositoryView)(nil)
	_ credentials.ConsumerIdentityProvider = (*repositoryView)(nil)
	_ utils.Unwrappable                    = (*repositoryView)(nil)
)

func GetRepositoryImplementation(n cpi2.Repository) (RepositoryImpl, error) {
	if v, ok := n.(*repositoryView); ok {
		return v.impl, nil
	}
	return nil, errors.ErrNotSupported("repository implementation type", fmt.Sprintf("%T", n))
}

func repositoryViewCreator(i RepositoryImpl, v resource.CloserView, d RepositoryViewManager) cpi2.Repository {
	return &repositoryView{
		_RepositoryView: resource.NewView[cpi2.Repository](v, d),
		impl:            i,
	}
}

// NewNoneRefRepositoryView provides a repository reflecting the state of the
// view manager without holding an additional reference.
func NewNoneRefRepositoryView(i RepositoryImpl) cpi2.Repository {
	return &repositoryView{
		_RepositoryView: resource.NewView[cpi2.Repository](resource.NewNonRefView[cpi2.Repository](i), i),
		impl:            i,
	}
}

func NewRepository(impl RepositoryImpl, name ...string) cpi2.Repository {
	return resource.NewResource[cpi2.Repository](impl, repositoryViewCreator, utils.OptionalDefaulted("OCM repo", name...), true)
}

func (r *repositoryView) Unwrap() interface{} {
	return r.impl
}

func (r *repositoryView) GetConsumerId(uctx ...credentials.UsageContext) credentials.ConsumerIdentity {
	return credentials.GetProvidedConsumerId(r.impl, uctx...)
}

func (r *repositoryView) GetIdentityMatcher() string {
	return credentials.GetProvidedIdentityMatcher(r.impl)
}

func (r *repositoryView) GetSpecification() cpi2.RepositorySpec {
	return r.impl.GetSpecification()
}

func (r *repositoryView) GetContext() cpi2.Context {
	return r.impl.GetContext()
}

func (r *repositoryView) ComponentLister() cpi2.ComponentLister {
	return r.impl.ComponentLister()
}

func (r *repositoryView) ExistsComponentVersion(name string, version string) (ok bool, err error) {
	err = r.Execute(func() error {
		ok, err = r.impl.ExistsComponentVersion(name, version)
		return err
	})
	return ok, err
}

func (r *repositoryView) LookupComponentVersion(name string, version string) (acc cpi2.ComponentVersionAccess, err error) {
	err = r.Execute(func() error {
		acc, err = r.impl.LookupComponentVersion(name, version)
		return err
	})
	return acc, err
}

func (r *repositoryView) LookupComponent(name string) (acc cpi2.ComponentAccess, err error) {
	err = r.Execute(func() error {
		acc, err = r.impl.LookupComponent(name)
		return err
	})
	return acc, err
}

func (r *repositoryView) NewComponentVersion(comp, vers string, overrides ...bool) (cpi2.ComponentVersionAccess, error) {
	c, err := refmgmt.ToLazy(r.LookupComponent(comp))
	if err != nil {
		return nil, err
	}
	defer c.Close()

	return c.NewVersion(vers, overrides...)
}

func (r *repositoryView) AddComponentVersion(cv cpi2.ComponentVersionAccess, overrides ...bool) error {
	c, err := refmgmt.ToLazy(r.LookupComponent(cv.GetName()))
	if err != nil {
		return err
	}
	defer c.Close()

	return c.AddVersion(cv, overrides...)
}
