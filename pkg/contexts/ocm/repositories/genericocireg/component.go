// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package genericocireg

import (
	"strings"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
)

type ComponentAccess struct {
	view accessio.CloserView // handle close and refs
	*componentAccessImpl
}

// implemented by view
// the rest is directly taken from the artefact set implementation

func (s *ComponentAccess) Close() error {
	return s.view.Close()
}

func (s *ComponentAccess) IsClosed() bool {
	return s.view.IsClosed()
}

////////////////////////////////////////////////////////////////////////////////

type componentAccessImpl struct {
	refs      accessio.ReferencableCloser
	repo      *Repository
	name      string
	namespace oci.NamespaceAccess
}

var _ cpi.ComponentAccess = (*ComponentAccess)(nil)

func newComponentAccess(repo *RepositoryImpl, name string, main bool) (*ComponentAccess, error) {
	mapped, err := repo.MapComponentNameToNamespace(name)
	if err != nil {
		return nil, err
	}
	v, err := repo.View(false)
	if err != nil {
		return nil, err
	}
	namespace, err := repo.ocirepo.LookupNamespace(mapped)
	if err != nil {
		v.Close()
		return nil, err
	}
	n := &componentAccessImpl{
		repo:      v,
		name:      name,
		namespace: namespace,
	}
	n.refs = accessio.NewRefCloser(n, true)
	return n.View(main)
}

func (a *componentAccessImpl) View(main ...bool) (*ComponentAccess, error) {
	v, err := a.refs.View(main...)
	if err != nil {
		return nil, err
	}
	return &ComponentAccess{view: v, componentAccessImpl: a}, nil
}

func (c *componentAccessImpl) GetName() string {
	return c.name
}

func (c *componentAccessImpl) Close() error {
	err := c.namespace.Close()
	if err != nil {
		c.repo.Close()
		return err
	}
	return c.repo.Close()
}

func (c *componentAccessImpl) GetContext() cpi.Context {
	return c.repo.GetContext()
}

////////////////////////////////////////////////////////////////////////////////

func (c *componentAccessImpl) ListVersions() ([]string, error) {
	tags, err := c.namespace.ListTags()
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(tags))
	for _, t := range tags {
		// omit reported digests (typically for ctf)
		if !strings.HasPrefix(t, "@") {
			result = append(result, t)
		}
	}
	return result, err
}

func (c *componentAccessImpl) LookupVersion(version string) (cpi.ComponentVersionAccess, error) {
	v, err := c.View(false)
	if err != nil {
		return nil, err
	}
	defer v.Close()
	acc, err := c.namespace.GetArtefact(version)
	if err != nil {
		if errors.IsErrNotFound(err) {
			return nil, cpi.ErrComponentVersionNotFoundWrap(err, c.name, version)
		}
		return nil, err
	}
	return newComponentVersionAccess(accessobj.ACC_WRITABLE, c, version, acc)
}

func (c *componentAccessImpl) AddVersion(access cpi.ComponentVersionAccess) error {
	if a, ok := access.(*ComponentVersion); ok {
		if a.GetName() != c.GetName() {
			return errors.ErrInvalid("component name", a.GetName())
		}
		return a.container.Update()
	}
	return errors.ErrInvalid("component version")
}

func (c *componentAccessImpl) NewVersion(version string, overrides ...bool) (cpi.ComponentVersionAccess, error) {
	v, err := c.View(false)
	if err != nil {
		return nil, err
	}
	defer v.Close()
	override := false
	for _, o := range overrides {
		if o {
			override = o
		}
	}
	acc, err := c.namespace.GetArtefact(version)
	if err == nil {
		if override {
			return newComponentVersionAccess(accessobj.ACC_CREATE, c, version, acc)
		}
		return nil, errors.ErrAlreadyExists(cpi.KIND_COMPONENTVERSION, c.name+"/"+version)
	}
	if !errors.IsErrNotFoundKind(err, oci.KIND_OCIARTEFACT) {
		return nil, err
	}
	acc, err = c.namespace.NewArtefact()
	if err != nil {
		return nil, err
	}
	return newComponentVersionAccess(accessobj.ACC_CREATE, c, version, acc)
}
