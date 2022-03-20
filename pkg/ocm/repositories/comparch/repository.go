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

package comparch

import (
	"strings"

	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/ocm/core"
	"github.com/gardener/ocm/pkg/ocm/cpi"
	impl "github.com/gardener/ocm/pkg/ocm/repositories/comparch/comparch"
)

type Repository struct {
	ctx  cpi.Context
	spec *RepositorySpec
	arch *impl.ComponentArchive
}

var _ cpi.Repository = (*Repository)(nil)

func NewRepository(ctx cpi.Context, s *RepositorySpec) (*Repository, error) {
	r := &Repository{ctx, s, nil}
	a, err := r.Open()
	if err != nil {
		return nil, err
	}
	r.arch = a
	return r, err
}

func (r *Repository) ComponentLister() cpi.ComponentLister {
	return r
}

func (r *Repository) NumComponents(prefix string) (int, error) {
	if r.arch == nil {
		return -1, accessio.ErrClosed
	}
	if r.arch.GetName() != prefix {
		if !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		if !strings.HasPrefix(r.arch.GetName(), prefix) {
			return 0, nil
		}
	}
	return 1, nil
}

func (r *Repository) GetComponents(prefix string, closure bool) ([]string, error) {
	if r.arch == nil {
		return nil, accessio.ErrClosed
	}
	if r.arch.GetName() != prefix {
		if !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		if !closure || !strings.HasPrefix(r.arch.GetName(), prefix) {
			return []string{}, nil
		}
	}
	return []string{r.arch.GetName()}, nil
}

func (r *Repository) Get() *impl.ComponentArchive {
	if r.arch != nil {
		return r.arch
	}
	return nil
}

func (r *Repository) Open() (*impl.ComponentArchive, error) {
	a, err := impl.Open(r.ctx, r.spec.AccessMode, r.spec.FilePath, 0700, r.spec.Options)
	if err != nil {
		return nil, err
	}
	r.arch = a
	return a, nil
}

func (r *Repository) GetContext() core.Context {
	return r.ctx
}

func (r *Repository) GetSpecification() core.RepositorySpec {
	return r.spec
}

func (r *Repository) ExistsComponentVersion(name string, ref string) (bool, error) {
	if r.arch == nil {
		return false, accessio.ErrClosed
	}
	return r.arch.GetName() == name && r.arch.GetVersion() == ref, nil
}

func (r *Repository) LookupComponentVersion(name string, version string) (cpi.ComponentVersionAccess, error) {
	ok, err := r.ExistsComponentVersion(name, version)
	if ok {
		return r.arch, nil
	}
	return nil, err
}

func (r *Repository) LookupComponent(name string) (cpi.ComponentAccess, error) {
	if r.arch == nil {
		return nil, accessio.ErrClosed
	}
	if r.arch.GetName() != name {
		return nil, errors.ErrNotFound(errors.KIND_COMPONENT, name, CTFComponentArchiveType)
	}
	return &ComponentAccess{r}, nil
}

func (r Repository) Close() error {
	if r.arch != nil {
		r.arch.Close()
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type ComponentAccess struct {
	repo *Repository
}

var _ cpi.ComponentAccess = (*ComponentAccess)(nil)

func (c *ComponentAccess) GetContext() cpi.Context {
	return c.repo.GetContext()
}

func (c *ComponentAccess) Close() error {
	return nil
}

func (c *ComponentAccess) GetName() string {
	return c.repo.arch.GetName()
}

func (c *ComponentAccess) ListVersions() ([]string, error) {
	return []string{c.repo.arch.GetVersion()}, nil
}

func (c *ComponentAccess) LookupVersion(ref string) (cpi.ComponentVersionAccess, error) {
	return c.repo.LookupComponentVersion(c.repo.arch.GetName(), ref)
}

func (c *ComponentAccess) AddVersion(access core.ComponentVersionAccess) error {
	return errors.ErrNotSupported(errors.KIND_FUNCTION, "add version", CTFComponentArchiveType)
}

func (c *ComponentAccess) NewVersion(version string) (core.ComponentVersionAccess, error) {
	return nil, errors.ErrNotSupported(errors.KIND_FUNCTION, "new version", CTFComponentArchiveType)
}
