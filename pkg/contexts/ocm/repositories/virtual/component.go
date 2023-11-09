// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package virtual

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi/repocpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/utils"
)

type componentAccessImpl struct {
	base repocpi.ComponentAccessBase

	repo *RepositoryImpl
	name string
}

var _ repocpi.ComponentAccessImpl = (*componentAccessImpl)(nil)

func newComponentAccess(repo *RepositoryImpl, name string, main bool) (cpi.ComponentAccess, error) {
	impl := &componentAccessImpl{
		repo: repo,
		name: name,
	}
	return repocpi.NewComponentAccess(impl, "OCM component[Simple]")
}

func (c *componentAccessImpl) Close() error {
	return nil
}

func (c *componentAccessImpl) SetBase(base repocpi.ComponentAccessBase) {
	c.base = base
}

func (c *componentAccessImpl) GetParentBase() repocpi.RepositoryViewManager {
	return c.repo.base
}

func (c *componentAccessImpl) GetContext() cpi.Context {
	return c.repo.GetContext()
}

func (c *componentAccessImpl) GetName() string {
	return c.name
}

func (c *componentAccessImpl) ListVersions() ([]string, error) {
	return c.repo.access.ListVersions(c.name)
}

func (c *componentAccessImpl) HasVersion(vers string) (bool, error) {
	return c.repo.ExistsComponentVersion(c.name, vers)
}

func (c *componentAccessImpl) IsReadOnly() bool {
	return c.repo.access.IsReadOnly()
}

func (c *componentAccessImpl) LookupVersion(version string) (cpi.ComponentVersionAccess, error) {
	ok, err := c.HasVersion(version)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, cpi.ErrComponentVersionNotFoundWrap(err, c.name, version)
	}
	v, err := c.base.View()
	if err != nil {
		return nil, err
	}
	defer v.Close()

	return newComponentVersionAccess(c, version, true)
}

func (c *componentAccessImpl) versionContainer(access cpi.ComponentVersionAccess) *ComponentVersionContainer {
	mine, _ := repocpi.GetComponentVersionImpl[*ComponentVersionContainer](access)
	if mine == nil || mine.comp != c {
		return nil
	}
	return mine
}

func (c *componentAccessImpl) NewVersion(version string, overrides ...bool) (cpi.ComponentVersionAccess, error) {
	v, err := c.base.View(false)
	if err != nil {
		return nil, err
	}
	defer v.Close()

	override := utils.Optional(overrides...)
	ok, err := c.HasVersion(version)
	if err == nil && ok {
		if override {
			return newComponentVersionAccess(c, version, false)
		}
		return nil, errors.ErrAlreadyExists(cpi.KIND_COMPONENTVERSION, c.name+"/"+version)
	}
	if err != nil && !errors.IsErrNotFoundKind(err, cpi.KIND_COMPONENTVERSION) {
		return nil, err
	}
	return newComponentVersionAccess(c, version, false)
}
