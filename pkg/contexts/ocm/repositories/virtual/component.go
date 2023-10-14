// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package virtual

import (
	"fmt"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi/support"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/utils"
)

type _ComponentAccessImplBase = cpi.ComponentAccessImplBase

type componentAccessImpl struct {
	_ComponentAccessImplBase
	repo *RepositoryImpl
	name string
}

func newComponentAccess(repo *RepositoryImpl, name string, main bool) (cpi.ComponentAccess, error) {
	base, err := cpi.NewComponentAccessImplBase(repo.GetContext(), name, repo)
	if err != nil {
		return nil, err
	}
	impl := &componentAccessImpl{
		_ComponentAccessImplBase: *base,
		repo:                     repo,
		name:                     name,
	}
	return cpi.NewComponentAccess(impl, "OCM component[Simple]"), nil
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
	v, err := c._ComponentAccessImplBase.View()
	if err != nil {
		return nil, err
	}
	defer v.Close()

	return newComponentVersionAccess(c, version, true)
}

func (c *componentAccessImpl) versionContainer(access cpi.ComponentVersionAccess) *ComponentVersionContainer {
	mine, _ := support.GetComponentVersionContainer[*ComponentVersionContainer](access)
	if mine == nil || mine.comp != c {
		return nil
	}
	return mine
}

func (c *componentAccessImpl) IsOwned(access cpi.ComponentVersionAccess) bool {
	return c.versionContainer(access) != nil
}

func (c *componentAccessImpl) AddVersion(access cpi.ComponentVersionAccess) error {
	if access.GetName() != c.GetName() {
		return errors.ErrInvalid("component name", access.GetName())
	}
	mine := c.versionContainer(access)
	if mine == nil {
		return fmt.Errorf("cannot add component version: component version access %s not created for target", access.GetName()+":"+access.GetVersion())
	}
	mine.impl.EnablePersistence()

	// delayed update in close is not done for composition mode,
	// so, we have to do the update, here.
	if !mine.impl.UseDirectAccess() || mine.impl.ShouldUpdate(false) {
		return mine.impl.Update()
	}
	return nil
}

func (c *componentAccessImpl) NewVersion(version string, overrides ...bool) (cpi.ComponentVersionAccess, error) {
	v, err := c.View(false)
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
