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
	"io"
	"sync"

	"github.com/gardener/ocm/pkg/errors"
)

type NamespaceContainer interface {
	LookupNamespace(name string) (NamespaceAccess, error)
}
type ArtefactContainer interface {
	GetArtefact(version string) (ArtefactAccess, error)
}

type Session interface {
	Closer(closer io.Closer, err error) (io.Closer, error)

	LookupNamespace(NamespaceContainer, string) (NamespaceAccess, error)
	GetArtefact(ArtefactContainer, string) (ArtefactAccess, error)

	Close() error
}

type objectkey struct {
	Object interface{}
	Name   string
}
type session struct {
	lock       sync.RWMutex
	closed     bool
	closer     []io.Closer
	namespaces map[objectkey]NamespaceAccess
	artefacts  map[objectkey]ArtefactAccess
}

func NewSession() Session {
	return &session{
		namespaces: map[objectkey]NamespaceAccess{},
		artefacts:  map[objectkey]ArtefactAccess{},
	}
}

func (s *session) Close() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	list := errors.ErrListf("closing session")
	for i := len(s.closer) - 1; i >= 0; i-- {
		list.Add(s.closer[i].Close())
	}
	s.namespaces = nil
	return list.Result()
}

func (s *session) Closer(closer io.Closer, err error) (io.Closer, error) {
	if err != nil {
		return nil, err
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	s.closer = append(s.closer, closer)
	return closer, err
}

func (s *session) add(closer io.Closer, err error) (io.Closer, error) {
	if err != nil {
		return nil, err
	}
	s.closer = append(s.closer, closer)
	return closer, err
}

func (s *session) LookupNamespace(c NamespaceContainer, name string) (NamespaceAccess, error) {
	key := objectkey{
		Object: c,
		Name:   name,
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.closed {
		return nil, errors.ErrClosed("session")
	}
	if ns := s.namespaces[key]; s != nil {
		return ns, nil
	}
	ns, err := c.LookupNamespace(name)
	if err != nil {
		return nil, err
	}
	s.namespaces[key] = ns
	s.add(ns, err)
	return ns, err
}

func (s *session) GetArtefact(c ArtefactContainer, version string) (ArtefactAccess, error) {
	key := objectkey{
		Object: c,
		Name:   version,
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.closed {
		return nil, errors.ErrClosed("session")
	}
	if obj := s.artefacts[key]; s != nil {
		return obj, nil
	}
	obj, err := c.GetArtefact(version)
	if err != nil {
		return nil, err
	}
	s.artefacts[key] = obj
	s.add(obj, err)
	return obj, err
}
