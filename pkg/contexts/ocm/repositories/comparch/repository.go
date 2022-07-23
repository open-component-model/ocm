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
	"sync"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/vfsattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
)

type Repository struct {
	lock sync.RWMutex
	ctx  cpi.Context
	spec *RepositorySpec
	arch *ComponentArchive
}

var _ cpi.Repository = (*Repository)(nil)

func NewRepository(ctx cpi.Context, s *RepositorySpec) (*Repository, error) {
	if s.PathFileSystem == nil {
		s.PathFileSystem = vfsattr.Get(ctx)
	}
	a, err := Open(ctx, s.AccessMode, s.FilePath, 0700, s)
	if err != nil {
		return nil, err
	}
	return a.comp.repo, nil
}

func (r *Repository) ComponentLister() cpi.ComponentLister {
	return r
}

func (r *Repository) matchPrefix(prefix string, closure bool) bool {
	if r.arch.GetName() != prefix {
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		if !closure || !strings.HasPrefix(r.arch.GetName(), prefix) {
			return false
		}
	}
	return true
}

func (r *Repository) NumComponents(prefix string) (int, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	if r.arch == nil {
		return -1, accessio.ErrClosed
	}
	if !r.matchPrefix(prefix, true) {
		return 0, nil
	}
	return 1, nil
}

func (r *Repository) GetComponents(prefix string, closure bool) ([]string, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	if r.arch == nil {
		return nil, accessio.ErrClosed
	}
	if !r.matchPrefix(prefix, closure) {
		return []string{}, nil
	}
	return []string{r.arch.GetName()}, nil
}

func (r *Repository) Get() *ComponentArchive {
	r.lock.RLock()
	defer r.lock.RUnlock()
	if r.arch != nil {
		return r.arch
	}
	return nil
}

func (r *Repository) Open() (*ComponentArchive, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.arch != nil {
		return r.arch, nil
	}
	a, err := Open(r.ctx, r.spec.AccessMode, r.spec.FilePath, 0700, r.spec)
	if err != nil {
		return nil, err
	}
	r.arch = a
	return a, nil
}

func (r *Repository) GetContext() cpi.Context {
	return r.ctx
}

func (r *Repository) GetSpecification() cpi.RepositorySpec {
	return r.spec
}

func (r *Repository) ExistsComponentVersion(name string, ref string) (bool, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	if r.arch == nil {
		return false, accessio.ErrClosed
	}
	return r.arch.GetName() == name && r.arch.GetVersion() == ref, nil
}

func (r *Repository) LookupComponentVersion(name string, version string) (cpi.ComponentVersionAccess, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	ok, err := r.ExistsComponentVersion(name, version)
	if ok {
		return r.arch, nil
	}
	if err == nil {
		err = errors.ErrNotFound(cpi.KIND_COMPONENTVERSION, common.NewNameVersion(name, version).String(), CTFComponentArchiveType)
	}
	return nil, err
}

func (r *Repository) LookupComponent(name string) (cpi.ComponentAccess, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	if r.arch == nil {
		return nil, accessio.ErrClosed
	}
	if r.arch.GetName() != name {
		return nil, errors.ErrNotFound(errors.KIND_COMPONENT, name, CTFComponentArchiveType)
	}
	return r.arch.comp, nil
}

func (r Repository) Close() error {
	r.lock.Lock()
	defer r.lock.Unlock()
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

func (c *ComponentAccess) AddVersion(access cpi.ComponentVersionAccess) error {
	return errors.ErrNotSupported(errors.KIND_FUNCTION, "add version", CTFComponentArchiveType)
}

func (c *ComponentAccess) NewVersion(version string, overrides ...bool) (cpi.ComponentVersionAccess, error) {
	return nil, errors.ErrNotSupported(errors.KIND_FUNCTION, "new version", CTFComponentArchiveType)
}
