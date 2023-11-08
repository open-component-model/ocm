// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package repocpi

import (
	"fmt"
	"io"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
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

////////////////////////////////////////////////////////////////////////////////

type _repositoryView interface {
	resource.ResourceViewInt[cpi.Repository] // here you have to redeclare
}

type RepositoryViewManager = resource.ViewManager[cpi.Repository] // here you have to use an alias

type RepositoryBase interface {
	resource.ResourceImplementation[cpi.Repository]

	GetContext() cpi.Context

	GetSpecification() cpi.RepositorySpec
	ComponentLister() cpi.ComponentLister

	ExistsComponentVersion(name string, version string) (bool, error)
	LookupComponentVersion(name string, version string) (cpi.ComponentVersionAccess, error)
	LookupComponent(name string) (cpi.ComponentAccess, error)

	io.Closer
}

type repositoryView struct {
	_repositoryView
	base RepositoryBase
}

var (
	_ cpi.Repository                       = (*repositoryView)(nil)
	_ credentials.ConsumerIdentityProvider = (*repositoryView)(nil)
	_ utils.Unwrappable                    = (*repositoryView)(nil)
)

func GetRepositoryBase(n cpi.Repository) (RepositoryBase, error) {
	if v, ok := n.(*repositoryView); ok {
		return v.base, nil
	}
	return nil, errors.ErrNotSupported("repository implementation type", fmt.Sprintf("%T", n))
}

func GetRepositoryImplementation(n cpi.Repository) (RepositoryImpl, error) {
	if v, ok := n.(*repositoryView); ok {
		if b, ok := v.base.(*repositoryBase); ok {
			return b.impl, nil
		}
		return nil, errors.ErrNotSupported("repository base type", fmt.Sprintf("%T", v.base))
	}
	return nil, errors.ErrNotSupported("repository implementation type", fmt.Sprintf("%T", n))
}

func repositoryViewCreator(i RepositoryBase, v resource.CloserView, d RepositoryViewManager) cpi.Repository {
	return &repositoryView{
		_repositoryView: resource.NewView[cpi.Repository](v, d),
		base:            i,
	}
}

// NewNoneRefRepositoryView provides a repository reflecting the state of the
// view manager without holding an additional reference.
func NewNoneRefRepositoryView(i RepositoryBase) cpi.Repository {
	return &repositoryView{
		_repositoryView: resource.NewView[cpi.Repository](resource.NewNonRefView[cpi.Repository](i), i),
		base:            i,
	}
}

func NewRepository(impl RepositoryImpl, kind string, closer ...io.Closer) cpi.Repository {
	base := newRepositoryImplBase(impl, kind, closer...)
	if kind == "" {
		kind = "OCM repository"
	}
	return resource.NewResource[cpi.Repository](base, repositoryViewCreator, kind, true)
}

func (r *repositoryView) Unwrap() interface{} {
	return r.base
}

func (r *repositoryView) GetConsumerId(uctx ...credentials.UsageContext) credentials.ConsumerIdentity {
	return credentials.GetProvidedConsumerId(r.base, uctx...)
}

func (r *repositoryView) GetIdentityMatcher() string {
	return credentials.GetProvidedIdentityMatcher(r.base)
}

func (r *repositoryView) GetSpecification() cpi.RepositorySpec {
	return r.base.GetSpecification()
}

func (r *repositoryView) GetContext() cpi.Context {
	return r.base.GetContext()
}

func (r *repositoryView) ComponentLister() cpi.ComponentLister {
	return r.base.ComponentLister()
}

func (r *repositoryView) ExistsComponentVersion(name string, version string) (ok bool, err error) {
	err = r.Execute(func() error {
		ok, err = r.base.ExistsComponentVersion(name, version)
		return err
	})
	return ok, err
}

func (r *repositoryView) LookupComponentVersion(name string, version string) (acc cpi.ComponentVersionAccess, err error) {
	err = r.Execute(func() error {
		acc, err = r.base.LookupComponentVersion(name, version)
		return err
	})
	return acc, err
}

func (r *repositoryView) LookupComponent(name string) (acc cpi.ComponentAccess, err error) {
	err = r.Execute(func() error {
		acc, err = r.base.LookupComponent(name)
		return err
	})
	return acc, err
}

func (r *repositoryView) NewComponentVersion(comp, vers string, overrides ...bool) (cpi.ComponentVersionAccess, error) {
	c, err := refmgmt.ToLazy(r.LookupComponent(comp))
	if err != nil {
		return nil, err
	}
	defer c.Close()

	return c.NewVersion(vers, overrides...)
}

func (r *repositoryView) AddComponentVersion(cv cpi.ComponentVersionAccess, overrides ...bool) error {
	c, err := refmgmt.ToLazy(r.LookupComponent(cv.GetName()))
	if err != nil {
		return err
	}
	defer c.Close()

	return c.AddVersion(cv, overrides...)
}
