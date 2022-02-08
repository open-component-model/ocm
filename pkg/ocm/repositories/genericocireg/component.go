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
	"github.com/gardener/ocm/pkg/common/accessobj"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci"
	"github.com/gardener/ocm/pkg/ocm/core"
	"github.com/gardener/ocm/pkg/ocm/cpi"
)

type ComponentAccess struct {
	repo      *Repository
	name      string
	namespace oci.NamespaceAccess
}

var _ cpi.ComponentAccess = (*ComponentAccess)(nil)

func NewComponentAccess(repo *Repository, name string) (*ComponentAccess, error) {
	mapped, err := repo.MapComponentNameToNamespace(name)
	if err != nil {
		return nil, err
	}
	namespace, err := repo.ocirepo.LookupNamespace(mapped)
	if err != nil {
		return nil, err
	}
	n := &ComponentAccess{
		repo:      repo,
		name:      name,
		namespace: namespace,
	}
	return n, err
}

func (c *ComponentAccess) GetName() string {
	return c.name
}

func (c *ComponentAccess) Close() error {
	return c.namespace.Close()
}

func (c *ComponentAccess) GetContext() core.Context {
	return c.repo.GetContext()
}

////////////////////////////////////////////////////////////////////////////////

func (c *ComponentAccess) LookupVersion(version string) (cpi.ComponentVersionAccess, error) {

	acc, err := c.namespace.GetArtefact(version)
	if err != nil {
		return nil, err
	}
	m := acc.ManifestAccess()
	if m == nil {
		return nil, errors.ErrInvalid("artefact type")
	}
	return NewComponentVersionAccess(accessobj.ACC_WRITABLE, c, version, m)
}

func (c *ComponentAccess) AddVersion(access cpi.ComponentVersionAccess) error {
	if a, ok := access.(*ComponentVersion); ok {
		if a.GetName() != c.GetName() {
			return errors.ErrInvalid("component name", a.GetName())
		}
		return a.container.Update()
	}
	return errors.ErrInvalid("component version")
}

func (c *ComponentAccess) NewVersion(version string) (cpi.ComponentVersionAccess, error) {
	_, err := c.namespace.GetArtefact(version)
	if err == nil {
		return nil, errors.ErrAlreadyExists(cpi.KIND_COMPONENTVERSION, c.name+"/"+version)
	}
	if !errors.IsErrNotFoundKind(err, oci.KIND_OCIARTEFACT) {
		return nil, err
	}
	acc, err := c.namespace.NewArtefact()
	if err != nil {
		return nil, err
	}
	return NewComponentVersionAccess(accessobj.ACC_CREATE, c, version, acc.ManifestAccess())
}
