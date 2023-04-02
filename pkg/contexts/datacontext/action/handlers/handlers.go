// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"sort"
	"sync"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/action"
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

type ActionHandlerMatch struct {
	Handler  ActionHandler
	Version  string
	Priority int
}

type Handlers interface {
	Register(kind string, versions []string, h ActionHandler, selectors ...api.Selector) error
	Execute(spec api.ActionSpec, creds common.Properties) (api.ActionResult, error)
	Get(spec api.ActionSpec, possible ...string) []ActionHandlerMatch
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
	result := r.Get(spec)
	sort.SliceStable(result, func(a, b int) bool {
		return result[a].Priority < result[b].Priority
	})
	if len(result) > 0 {
		spec.SetType(runtime.TypeName(spec.GetKind(), result[0].Version))
		return result[0].Handler.Handle(spec, creds)
	}
	return nil, nil
}

func (r *registry) Get(spec api.ActionSpec, possible ...string) []ActionHandlerMatch {
	if len(possible) == 0 {
		possible = api.SupportedActionVersions(spec.GetKind())
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	var result []ActionHandlerMatch

	if kinds := r.registrations[spec.GetKind()]; kinds != nil {
		if reg := kinds[spec.Selector()]; reg != nil {
			if len(reg.versions) != 0 {
				if v := MatchVersion(action.SupportedActionVersions(spec.GetKind()), reg.versions); v != "" {
					result = append(result, ActionHandlerMatch{Handler: reg.handler, Version: v, Priority: 0})
				}
			}
		}
	}

	if r.base != nil {
		result = append(result, r.base.Get(spec, possible...)...)
	}
	return result
}

func MatchVersion(possible []string, avail []string) string {
	p := append(possible[:0:0], possible...) //nolint: gocritic // yes
	a := append(avail[:0:0], avail...)       //nolint: gocritic // yes

	scheme.SortVersions(p)
	scheme.SortVersions(a)
	f := ""
	for _, v := range p {
		for _, c := range a {
			if v == c {
				f = c
				break
			}
		}
	}
	return f
}
