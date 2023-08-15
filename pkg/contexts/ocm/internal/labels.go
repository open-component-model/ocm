// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"sync"

	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/utils"
)

type LabelMergeHandlerConfig interface {
	Complete(ctx Context) error
}

type LabelMergeHandler interface {
	Algorithm() string
	Description() string
	DecodeConfig(data []byte) (LabelMergeHandlerConfig, error)

	Merge(ctx Context, src, tgt *metav1.Label, cfg LabelMergeHandlerConfig) error
}

type LabelMergeHandlerRegistry interface {
	RegisterHandler(h LabelMergeHandler)
	AssignHandler(name string, typ string)

	GetHandler(name string) LabelMergeHandler
	GetHandlerFor(labeltyp string) LabelMergeHandler
	GetAlgorithmFor(typ string) string

	GetAlgorithms() Algorithms
	GetAssignments() Assignments

	Copy() LabelMergeHandlerRegistry
}

type mergeHandlerRegistry struct {
	lock sync.Mutex
	base LabelMergeHandlerRegistry

	handlerTypes map[string]LabelMergeHandler
	assignments  map[string]string
}

var _ LabelMergeHandlerRegistry = (*mergeHandlerRegistry)(nil)

func NewLabelMergeHandlerRegistry(base ...LabelMergeHandlerRegistry) LabelMergeHandlerRegistry {
	return &mergeHandlerRegistry{
		base:         utils.Optional(base...),
		handlerTypes: map[string]LabelMergeHandler{},
		assignments:  map[string]string{},
	}
}

func (m *mergeHandlerRegistry) RegisterHandler(h LabelMergeHandler) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.handlerTypes[h.Algorithm()] = h
}

func (m *mergeHandlerRegistry) AssignHandler(algo string, typ string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.assignments[typ] = algo
}

func (m *mergeHandlerRegistry) GetHandler(algo string) LabelMergeHandler {
	m.lock.Lock()
	defer m.lock.Unlock()
	h := m.handlerTypes[algo]
	if h == nil && m.base != nil {
		return m.base.GetHandler(algo)
	}
	return h
}

func (m *mergeHandlerRegistry) GetAlgorithmFor(typ string) string {
	m.lock.Lock()
	defer m.lock.Unlock()
	h := m.assignments[typ]
	if h == "" {
		k, _ := runtime.KindVersion(typ)
		h = m.assignments[k]
	}
	if h == "" && m.base != nil {
		return m.base.GetAlgorithmFor(typ)
	}
	return h
}

func (m *mergeHandlerRegistry) GetHandlerFor(typ string) LabelMergeHandler {
	n := m.GetAlgorithmFor(typ)
	if n == "" {
		return nil
	}
	return m.GetHandler(n)
}

func (m *mergeHandlerRegistry) Copy() LabelMergeHandlerRegistry {
	m.lock.Lock()
	defer m.lock.Unlock()
	c := &mergeHandlerRegistry{
		base:         m.base,
		handlerTypes: map[string]LabelMergeHandler{},
		assignments:  map[string]string{},
	}
	for k, v := range m.handlerTypes {
		c.handlerTypes[k] = v
	}
	for k, v := range m.assignments {
		c.assignments[k] = v
	}
	return c
}

type Algorithms = map[string]LabelMergeHandler

func (m *mergeHandlerRegistry) GetAlgorithms() Algorithms {
	m.lock.Lock()
	defer m.lock.Unlock()

	r := Algorithms{}
	if m.base != nil {
		r = m.base.GetAlgorithms()
	}
	for k, v := range m.handlerTypes {
		r[k] = v
	}
	return r
}

type Assignments = map[string]string

func (m *mergeHandlerRegistry) GetAssignments() Assignments {
	m.lock.Lock()
	defer m.lock.Unlock()

	r := Assignments{}
	if m.base != nil {
		r = m.base.GetAssignments()
	}
	for k, v := range m.assignments {
		r[k] = v
	}
	return r
}

////////////////////////////////////////////////////////////////////////////////

var DefaultLabelMergeHandlerRegistry = NewLabelMergeHandlerRegistry()
