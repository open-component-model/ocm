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

package oci

import (
	"encoding/json"
	"reflect"

	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/errors"
)

type NamespaceContainer interface {
	LookupNamespace(name string) (NamespaceAccess, error)
}
type ArtefactContainer interface {
	GetArtefact(version string) (ArtefactAccess, error)
}

type EvaluationResult struct {
	Ref        RefSpec
	Repository Repository
	Namespace  NamespaceAccess
	Artefact   ArtefactAccess
}

type Session interface {
	datacontext.Session

	LookupRepository(Context, RepositorySpec) (Repository, error)
	LookupNamespace(NamespaceContainer, string) (NamespaceAccess, error)
	GetArtefact(ArtefactContainer, string) (ArtefactAccess, error)
	EvaluateRef(ctx Context, ref string) (*EvaluationResult, error)
	DetermineRepository(ctx Context, ref string) (Repository, UniformRepositorySpec, error)
	DetermineRepositoryBySpec(ctx Context, spec *UniformRepositorySpec) (Repository, error)
}

type session struct {
	datacontext.Session
	base         datacontext.SessionBase
	repositories map[datacontext.ObjectKey]Repository
	namespaces   map[datacontext.ObjectKey]NamespaceAccess
	artefacts    map[datacontext.ObjectKey]ArtefactAccess
}

var key = reflect.TypeOf(session{})

func NewSession(s datacontext.Session) Session {
	return datacontext.GetOrCreateSubSession(s, key, newSession).(Session)
}

func newSession(s datacontext.SessionBase) datacontext.Session {
	return &session{
		Session:      s.Session(),
		base:         s,
		repositories: map[datacontext.ObjectKey]Repository{},
		namespaces:   map[datacontext.ObjectKey]NamespaceAccess{},
		artefacts:    map[datacontext.ObjectKey]ArtefactAccess{},
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

func (s *session) LookupNamespace(c NamespaceContainer, name string) (NamespaceAccess, error) {
	key := datacontext.ObjectKey{
		Object: c,
		Name:   name,
	}
	s.base.Lock()
	defer s.base.Unlock()
	if s.base.IsClosed() {
		return nil, errors.ErrClosed("session")
	}
	if ns := s.namespaces[key]; ns != nil {
		return ns, nil
	}
	ns, err := c.LookupNamespace(name)
	if err != nil {
		return nil, err
	}
	s.namespaces[key] = ns
	s.base.AddCloser(ns)
	return ns, err
}

func (s *session) GetArtefact(c ArtefactContainer, version string) (ArtefactAccess, error) {
	key := datacontext.ObjectKey{
		Object: c,
		Name:   version,
	}
	s.base.Lock()
	defer s.base.Unlock()
	if s.base.IsClosed() {
		return nil, errors.ErrClosed("session")
	}
	if obj := s.artefacts[key]; obj != nil {
		return obj, nil
	}
	obj, err := c.GetArtefact(version)
	if err != nil {
		return nil, err
	}
	s.artefacts[key] = obj
	s.base.AddCloser(obj)
	return obj, err
}

func (s *session) EvaluateRef(ctx Context, ref string) (*EvaluationResult, error) {
	var err error
	result := &EvaluationResult{}
	result.Ref, err = ParseRef(ref)
	if err != nil {
		return nil, err
	}
	result.Repository, err = s.DetermineRepositoryBySpec(ctx, &result.Ref.UniformRepositorySpec)
	if err != nil {
		return nil, err
	}
	result.Namespace, err = s.LookupNamespace(result.Repository, result.Ref.Repository)

	if !result.Ref.IsVersion() {
		return result, err
	}
	result.Artefact, err = s.GetArtefact(result.Namespace, result.Ref.Version())
	return result, err
}

func (s *session) DetermineRepository(ctx Context, ref string) (Repository, UniformRepositorySpec, error) {
	spec, err := ParseRepo(ref)
	if err != nil {
		return nil, spec, err
	}
	r, err := s.DetermineRepositoryBySpec(ctx, &spec)
	return r, spec, err
}

func (s *session) DetermineRepositoryBySpec(ctx Context, spec *UniformRepositorySpec) (Repository, error) {
	rspec, err := ctx.MapUniformRepositorySpec(spec)
	if err != nil {
		return nil, err
	}
	return s.LookupRepository(ctx, rspec)
}
