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

package ocm

import (
	"encoding/json"
	"reflect"

	"github.com/gardener/ocm/pkg/datacontext"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/ocm/repositories/genericocireg"
	"github.com/gardener/ocm/pkg/ocm/repositories/ocireg"
)

type ComponentContainer interface {
	LookupComponent(name string) (ComponentAccess, error)
}
type ComponentVersionContainer interface {
	LookupVersion(version string) (ComponentVersionAccess, error)
}

type Session interface {
	datacontext.Session

	LookupRepository(Context, RepositorySpec) (Repository, error)
	LookupComponent(ComponentContainer, string) (ComponentAccess, error)
	GetComponentVersion(ComponentVersionContainer, string) (ComponentVersionAccess, error)
	EvaluateRef(ctx Context, ref string) (*RefSpec, ComponentAccess, ComponentVersionAccess, error)
	Close() error
}

type session struct {
	datacontext.Session
	base         datacontext.SessionBase
	repositories map[datacontext.ObjectKey]Repository
	components   map[datacontext.ObjectKey]ComponentAccess
	versions     map[datacontext.ObjectKey]ComponentVersionAccess
}

var _ Session = (*session)(nil)

var key = reflect.TypeOf(session{})

func NewSession(s datacontext.Session) Session {
	return datacontext.GetOrCreateSubSession(s, key, newSession).(Session)
}

func newSession(s datacontext.SessionBase) datacontext.Session {
	return &session{
		Session:      s.Session(),
		base:         s,
		repositories: map[datacontext.ObjectKey]Repository{},
		components:   map[datacontext.ObjectKey]ComponentAccess{},
		versions:     map[datacontext.ObjectKey]ComponentVersionAccess{},
	}
}

func (s *session) LookupRepository(ctx Context, spec RepositorySpec) (Repository, error) {

	spec, err := ctx.RepositoryTypes().CreateRepositorySpec(spec)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}
	key := datacontext.ObjectKey{
		Object: ctx,
		Name:   string(data),
	}

	s.base.Lock()
	defer s.base.Unlock()
	if s.base.IsClosed() {
		return nil, errors.ErrClosed("session")
	}

	if r := s.repositories[key]; r != nil {
		return r, nil
	}
	repo, err := ctx.RepositoryForSpec(spec)
	if err != nil {
		return nil, err
	}
	s.repositories[key] = repo
	s.base.AddCloser(repo)
	return repo, err
}

func (s *session) LookupComponent(c ComponentContainer, name string) (ComponentAccess, error) {
	key := datacontext.ObjectKey{
		Object: c,
		Name:   name,
	}
	s.base.Lock()
	defer s.base.Unlock()
	if s.base.IsClosed() {
		return nil, errors.ErrClosed("session")
	}
	if ns := s.components[key]; ns != nil {
		return ns, nil
	}
	ns, err := c.LookupComponent(name)
	if err != nil {
		return nil, err
	}
	s.components[key] = ns
	s.base.AddCloser(ns)
	return ns, err
}

func (s *session) GetComponentVersion(c ComponentVersionContainer, version string) (ComponentVersionAccess, error) {
	key := datacontext.ObjectKey{
		Object: c,
		Name:   version,
	}
	s.base.Lock()
	defer s.base.Unlock()
	if s.base.IsClosed() {
		return nil, errors.ErrClosed("session")
	}
	if obj := s.versions[key]; s != nil {
		return obj, nil
	}
	obj, err := c.LookupVersion(version)
	if err != nil {
		return nil, err
	}
	s.versions[key] = obj
	s.base.AddCloser(obj)
	return obj, err
}

func (s *session) EvaluateRef(ctx Context, ref string) (*RefSpec, ComponentAccess, ComponentVersionAccess, error) {
	parsed, err := ParseRef(ref)
	if err != nil {
		return nil, nil, nil, err
	}
	meta := genericocireg.NewComponentRepositoryMeta(parsed.SubPath, "")
	spec := ocireg.NewRepositorySpec(parsed.Host, meta)
	repo, err := s.LookupRepository(ctx, spec)
	if err != nil {
		return nil, nil, nil, err
	}
	ns, err := s.LookupComponent(repo, parsed.Component)
	if !parsed.IsVersion() {
		return &parsed, ns, nil, err
	}
	v, err := s.GetComponentVersion(ns, *parsed.Version)
	return &parsed, ns, v, err
}
