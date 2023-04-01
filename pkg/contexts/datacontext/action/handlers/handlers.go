// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"sync"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/action/api"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/runtime/scheme"
)

var defaultHandlers = NewHandlers(nil)

func DefaultHandlers() Handlers {
	return defaultHandlers
}

type ActionHandler interface {
	Handle(api.ActionSpec, common.Properties) (api.ActionResult, error)
}

type Handlers interface {
	Register(kind string, versions []string, h ActionHandler, selectors ...api.Selector) error
	Execute(spec api.ActionSpec, creds common.Properties) (api.ActionResult, error)
	AddTo(t Handlers)
}

type registration struct {
	handler  ActionHandler
	versions []string
}

type registry struct {
	lock          sync.Mutex
	base          Handlers
	registrations map[string]map[api.Selector]*registration
}

var _ Handlers = (*registry)(nil)

func NewHandlers(base Handlers) Handlers {
	return &registry{
		base:          base,
		registrations: map[string]map[api.Selector]*registration{},
	}
}

func (r *registry) AddTo(t Handlers) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.base != nil {
		r.base.AddTo(t)
	}
	for k, sel := range r.registrations {
		for s, reg := range sel {
			t.Register(k, reg.versions, reg.handler, s)
		}
	}
}

func (r *registry) Register(kind string, versions []string, h ActionHandler, selectors ...api.Selector) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	kinds := r.registrations[kind]
	if kinds == nil {
		kinds = map[api.Selector]*registration{}
		r.registrations[kind] = kinds
	}

	versions = append(versions[:0:0], versions...)
	scheme.SortVersions(versions)
	reg := &registration{
		handler:  h,
		versions: versions,
	}

	for _, s := range selectors {
		kinds[s] = reg
	}
	return nil
}

func (r *registry) Execute(spec api.ActionSpec, creds common.Properties) (api.ActionResult, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	kinds := r.registrations[spec.GetKind()]
	if kinds == nil {
		return r.delegate(spec, creds)
	}
	reg := kinds[spec.Selector()]
	if reg == nil {
		return r.delegate(spec, creds)
	}
	if len(reg.versions) == 0 {
		return r.delegate(spec, creds)
	}
	spec.SetType(runtime.TypeName(spec.GetKind(), reg.versions[len(reg.versions)-1]))
	return reg.handler.Handle(spec, creds)
}

func (r *registry) delegate(spec api.ActionSpec, creds common.Properties) (api.ActionResult, error) {
	if r.base == nil {
		return nil, nil
	}
	return r.base.Execute(spec, creds)
}
