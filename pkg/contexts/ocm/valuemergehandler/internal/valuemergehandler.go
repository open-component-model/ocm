// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"sync"

	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/utils"
)

type Config interface {
	Complete(ctx Context) error
}

type Handler interface {
	Algorithm() string
	Description() string
	DecodeConfig(data []byte) (Config, error)

	Merge(ctx Context, src runtime.RawValue, tgt *runtime.RawValue, cfg Config) (bool, error)
}

type Registry interface {
	RegisterHandler(h Handler)
	AssignHandler(hint string, spec *Specification)

	GetHandler(name string) Handler
	GetAssignment(typ string) *Specification

	GetHandlers() Handlers
	GetAssignments() MergeHandlerAssignments

	Copy() Registry
}

type registry struct {
	lock sync.Mutex
	base Registry

	handlerTypes Handlers
	assignments  MergeHandlerAssignments
}

var _ Registry = (*registry)(nil)

func NewRegistry(base ...Registry) Registry {
	return &registry{
		base:         utils.Optional(base...),
		handlerTypes: Handlers{},
		assignments:  MergeHandlerAssignments{},
	}
}

func (m *registry) RegisterHandler(h Handler) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.handlerTypes[h.Algorithm()] = h
}

func (m *registry) AssignHandler(hint string, spec *Specification) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.assignments[hint] = spec
}

func (m *registry) GetHandler(algo string) Handler {
	m.lock.Lock()
	defer m.lock.Unlock()
	h := m.handlerTypes[algo]
	if h == nil && m.base != nil {
		return m.base.GetHandler(algo)
	}
	return h
}

func (m *registry) GetAssignment(hint string) *Specification {
	m.lock.Lock()
	defer m.lock.Unlock()
	h := m.assignments[hint]
	if h == nil && m.base != nil {
		return m.base.GetAssignment(hint)
	}
	return h
}

func (m *registry) Copy() Registry {
	m.lock.Lock()
	defer m.lock.Unlock()
	c := &registry{
		base:         m.base,
		handlerTypes: map[string]Handler{},
		assignments:  map[string]*Specification{},
	}
	for k, v := range m.handlerTypes {
		c.handlerTypes[k] = v
	}
	for k, v := range m.assignments {
		c.assignments[k] = v
	}
	return c
}

type Handlers = map[string]Handler

func (m *registry) GetHandlers() Handlers {
	m.lock.Lock()
	defer m.lock.Unlock()

	r := Handlers{}
	if m.base != nil {
		r = m.base.GetHandlers()
	}
	for k, v := range m.handlerTypes {
		r[k] = v
	}
	return r
}

type MergeHandlerAssignments = map[string]*Specification

func (m *registry) GetAssignments() MergeHandlerAssignments {
	m.lock.Lock()
	defer m.lock.Unlock()

	r := MergeHandlerAssignments{}
	if m.base != nil {
		r = m.base.GetAssignments()
	}
	for k, v := range m.assignments {
		r[k] = v
	}
	return r
}

////////////////////////////////////////////////////////////////////////////////

var DefaultRegistry = NewRegistry()
