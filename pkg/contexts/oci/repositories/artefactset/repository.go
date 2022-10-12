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

package artefactset

import (
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/vfsattr"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
)

type Repository struct {
	ctx  cpi.Context
	spec *RepositorySpec
	arch *ArtefactSet
}

var _ cpi.Repository = (*Repository)(nil)

func NewRepository(ctx cpi.Context, s *RepositorySpec) (*Repository, error) {
	if s.PathFileSystem == nil {
		s.PathFileSystem = vfsattr.Get(ctx)
	}
	r := &Repository{ctx, s, nil}
	_, err := r.Open()
	if err != nil {
		return nil, err
	}
	return r, err
}

func (r *Repository) Get() *ArtefactSet {
	if r.arch != nil {
		return r.arch
	}
	return nil
}

func (r *Repository) Open() (*ArtefactSet, error) {
	a, err := Open(r.spec.AccessMode, r.spec.FilePath, 0o700, &Options{}, &r.spec.Options, accessio.PathFileSystem(r.spec.PathFileSystem))
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

func (r *Repository) NamespaceLister() cpi.NamespaceLister {
	return anonymous
}

func (r *Repository) ExistsArtefact(name string, ref string) (bool, error) {
	if name != "" {
		return false, nil
	}
	return r.arch.HasArtefact(ref)
}

func (r *Repository) LookupArtefact(name string, ref string) (cpi.ArtefactAccess, error) {
	if name != "" {
		return nil, cpi.ErrUnknownArtefact(name, ref)
	}
	return r.arch.GetArtefact(ref)
}

func (r *Repository) LookupNamespace(name string) (cpi.NamespaceAccess, error) {
	if name != "" {
		return nil, errors.ErrNotSupported("namespace", name)
	}
	return r.arch, nil
}

func (r Repository) Close() error {
	if r.arch != nil {
		r.arch.Close()
	}
	return nil
}

// NamespaceLister handles the namespaces provided by an artefact set.
// This is always single anonymous namespace, which by ddefinition
// is the empty string.
type NamespaceLister struct{}

var anonymous cpi.NamespaceLister = &NamespaceLister{}

// NumNamespaces returns the number of namespaces with a given prefix
// for an artefact set. This is either one (the anonymous namespace) if
// the prefix is empty (all namespaces) or zero if a prefix is given.
func (n *NamespaceLister) NumNamespaces(prefix string) (int, error) {
	if prefix == "" {
		return 1, nil
	}
	return 0, nil
}

// GetNamespaces returns namespaces with a given prefix.
// This is the anonymous namespace ("") for an empty prefix
// or no namespace at all if a prefix is given.
func (n *NamespaceLister) GetNamespaces(prefix string, closure bool) ([]string, error) {
	if prefix == "" {
		return []string{""}, nil
	}
	return nil, nil
}
