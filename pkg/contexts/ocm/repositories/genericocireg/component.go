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

	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
)

type ComponentAccess struct {
	repo      *Repository
	name      string
	namespace oci.NamespaceAccess
	priv      bool // private access for dedicated component version
}

var _ cpi.ComponentAccess = (*ComponentAccess)(nil)

func newComponentAccess(repo *Repository, name string, priv bool) (*ComponentAccess, error) {
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
		priv:      priv,
	}
	return n, err
}

func (c *ComponentAccess) GetName() string {
	return c.name
}

func (c *ComponentAccess) Close() error {
	if !c.priv {
		c.repo.Close()
	}
	return c.namespace.Close()
}

func (c *ComponentAccess) GetContext() cpi.Context {
	return c.repo.GetContext()
}

////////////////////////////////////////////////////////////////////////////////

func (c *ComponentAccess) ListVersions() ([]string, error) {
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

func (c *ComponentAccess) LookupVersion(version string) (cpi.ComponentVersionAccess, error) {

	acc, err := c.namespace.GetArtefact(version)
	if err != nil {
		if errors.IsErrNotFound(err) {
			return nil, cpi.ErrComponentVersionNotFoundWrap(err, c.name, version)
		}
		return nil, err
	}
	m := acc.ManifestAccess()
	if m == nil {
		acc.Close()
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

func (c *ComponentAccess) NewVersion(version string, overrides ...bool) (cpi.ComponentVersionAccess, error) {
	override := false
	for _, o := range overrides {
		if o {
			override = o
		}
	}
	acc, err := c.namespace.GetArtefact(version)
	if err == nil {
		if override {
			return NewComponentVersionAccess(accessobj.ACC_CREATE, c, version, acc.ManifestAccess())
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
	return NewComponentVersionAccess(accessobj.ACC_CREATE, c, version, acc.ManifestAccess())
}
