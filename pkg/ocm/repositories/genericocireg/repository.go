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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path"
	"strings"

	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/oci"
	"github.com/open-component-model/ocm/pkg/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/ocm/repositories/genericocireg/componentmapping"
)

type Repository struct {
	ctx     cpi.Context
	meta    ComponentRepositoryMeta
	ocirepo oci.Repository
}

var _ cpi.Repository = (*Repository)(nil)

func NewRepository(ctx cpi.Context, meta *ComponentRepositoryMeta, ocirepo oci.Repository) (cpi.Repository, error) {
	repo := &Repository{
		ctx:     ctx,
		meta:    *DefaultComponentRepositoryMeta(meta),
		ocirepo: ocirepo,
	}
	_ = repo
	return repo, nil
}

func (r *Repository) Close() error {
	return r.ocirepo.Close()
}

func (r *Repository) GetContext() cpi.Context {
	return r.ctx
}

func (r *Repository) GetSpecification() cpi.RepositorySpec {
	return &RepositorySpec{
		RepositorySpec:          r.ocirepo.GetSpecification(),
		ComponentRepositoryMeta: r.meta,
	}
}

func (r *Repository) ComponentLister() cpi.ComponentLister {
	if r.meta.ComponentNameMapping != OCIRegistryURLPathMapping {
		return nil
	}
	lister := r.ocirepo.NamespaceLister()
	if lister == nil {
		return nil
	}
	return r
}

func (r *Repository) NumComponents(prefix string) (int, error) {
	lister := r.ocirepo.NamespaceLister()
	if lister == nil {
		return -1, errors.ErrNotSupported("component lister")
	}
	p := path.Join(r.meta.SubPath, componentmapping.ComponentDescriptorNamespace, prefix)
	if strings.HasSuffix(prefix, "/") && !strings.HasSuffix(p, "/") {
		p = p + "/"
	}
	return lister.NumNamespaces(p)
}

func (r *Repository) GetComponents(prefix string, closure bool) ([]string, error) {
	lister := r.ocirepo.NamespaceLister()
	if lister == nil {
		return nil, errors.ErrNotSupported("component lister")
	}
	p := path.Join(r.meta.SubPath, componentmapping.ComponentDescriptorNamespace)
	compprefix := len(p) + 1
	p = path.Join(p, prefix)
	if strings.HasSuffix(prefix, "/") && !strings.HasSuffix(p, "/") {
		p = p + "/"
	}
	tmp, err := lister.GetNamespaces(p, closure)
	if err != nil {
		return nil, err
	}
	result := make([]string, len(tmp))
	for i, r := range tmp {
		result[i] = r[compprefix:]
	}
	return result, nil
}

func (r *Repository) GetOCIRepository() oci.Repository {
	return r.ocirepo
}

func (r *Repository) ExistsComponentVersion(name string, version string) (bool, error) {
	namespace, err := r.MapComponentNameToNamespace(name)
	if err != nil {
		return false, err
	}
	ns, err := r.ocirepo.LookupNamespace(namespace)
	if err != nil {
		return false, err
	}
	a, err := ns.GetArtefact(version)
	if err != nil {
		return false, err
	}
	desc, err := a.Manifest()
	if err != nil {
		return false, err
	}
	switch desc.Config.MediaType {
	case componentmapping.ComponentDescriptorConfigMimeType, componentmapping.ComponentDescriptorLegacyConfigMimeType:
		return true, nil
	}
	return false, nil
}

func (r *Repository) LookupComponent(name string) (cpi.ComponentAccess, error) {
	return NewComponentAccess(r, name)
}

func (r *Repository) LookupComponentVersion(name string, version string) (cpi.ComponentVersionAccess, error) {
	c, err := r.LookupComponent(name)
	if err != nil {
		return nil, err
	}
	return c.LookupVersion(version)
}

func (r *Repository) MapComponentNameToNamespace(name string) (string, error) {
	switch r.meta.ComponentNameMapping {
	case OCIRegistryURLPathMapping, "":
		return path.Join(r.meta.SubPath, componentmapping.ComponentDescriptorNamespace, name), nil
	case OCIRegistryDigestMapping:
		h := sha256.New()
		_, _ = h.Write([]byte(name))
		return path.Join(r.meta.SubPath, hex.EncodeToString(h.Sum(nil))), nil
	default:
		return "", fmt.Errorf("unknown component name mapping method %s", r.meta.ComponentNameMapping)
	}
}
