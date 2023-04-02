// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/action"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/action/api"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/generics"
	"github.com/open-component-model/ocm/pkg/registrations"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/runtime/scheme"
	"github.com/open-component-model/ocm/pkg/utils"
)

var defaultHandlers = NewRegistry()

func DefaultRegistry() Registry {
	return defaultHandlers
}

////////////////////////////////////////////////////////////////////////////////

type ActionsProvider interface {
	GetActions() Registry
}

type ActionHandler interface {
	Handle(api.ActionSpec, common.Properties) (api.ActionResult, error)
}

type ActionHandlerMatch struct {
	Handler  ActionHandler
	Version  string
	Priority int
}

type (
	Target                      = ActionsProvider
	HandlerConfig               = registrations.HandlerConfig
	HandlerRegistrationHandler  = registrations.HandlerRegistrationHandler[Target, Option]
	HandlerRegistrationRegistry = registrations.HandlerRegistrationRegistry[Target, Option]
)

func NewHandlerRegistrationRegistry(base ...HandlerRegistrationRegistry) HandlerRegistrationRegistry {
	return registrations.NewHandlerRegistrationRegistry[Target, Option](base...)
}

type Registry interface {
	registrations.HandlerRegistrationRegistry[Target, Option]

	Register(h ActionHandler, opts ...Option) error
	Execute(spec api.ActionSpec, creds common.Properties) (api.ActionResult, error)
	Get(spec api.ActionSpec, possible ...string) []ActionHandlerMatch
	AddTo(t Registry)
}

type registration struct {
	handler  ActionHandler
	versions []string
	priority int
}

var _ Option = (*registration)(nil)

func (r *registration) ApplyActionHandlerOptionTo(opts *api.Options) {
	opts.Priority = r.priority
}

type registry struct {
	registrations.HandlerRegistrationRegistry[Target, Option]

	lock          sync.Mutex
	base          Registry
	registrations map[string]map[api.Selector]*registration
}

var _ Registry = (*registry)(nil)

func NewRegistry(base ...Registry) Registry {
	b := utils.Optional(base...)
	r := &registry{
		base:                        b,
		registrations:               map[string]map[api.Selector]*registration{},
		HandlerRegistrationRegistry: NewHandlerRegistrationRegistry(b),
	}
	return r
}

func (r *registry) AddTo(t Registry) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.base != nil {
		r.base.AddTo(t)
	}
	for k, sel := range r.registrations {
		for s, reg := range sel {
			t.Register(reg.handler, ForAction(k), WithVersions(reg.versions...), s, reg)
		}
	}
}

func (r *registry) Register(h ActionHandler, olist ...Option) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	opts := NewOptions(olist...)
	if opts.Action == "" {
		return fmt.Errorf("action kind required for action handler registration")
	}

	kinds := r.registrations[opts.Action]
	if kinds == nil {
		kinds = map[api.Selector]*registration{}
		r.registrations[opts.Action] = kinds
	}

	versions := opts.Versions
	if versions == nil {
		versions = action.SupportedActionVersions(opts.Action)
	}
	versions = append(versions[:0:0], versions...)
	if err := scheme.SortVersions(versions); err != nil {
		return errors.Wrapf(err, "invalid version set")
	}
	reg := &registration{
		handler:  h,
		versions: versions,
		priority: generics.Conditional(opts.Priority >= 0, opts.Priority, 10),
	}

	for _, s := range opts.Selectors {
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
		// first, check direct selctor match
		if reg := kinds[spec.Selector()]; reg != nil {
			if len(reg.versions) != 0 {
				if v := MatchVersion(action.SupportedActionVersions(spec.GetKind()), reg.versions); v != "" {
					result = append(result, ActionHandlerMatch{Handler: reg.handler, Version: v, Priority: reg.priority})
				}
			}
		} else {
			// second, try registrations as regexp matcher
			for sel, reg := range kinds {
				s := string(sel)
				e, err := regexp.Compile(s)
				if err == nil {
					t := strings.Trim(s, "^$")
					if t == s {
						e, err = regexp.Compile("^" + s + "$")
					}
				}
				if err == nil {
					if e.MatchString(string(spec.Selector())) {
						if v := MatchVersion(action.SupportedActionVersions(spec.GetKind()), reg.versions); v != "" {
							result = append(result, ActionHandlerMatch{Handler: reg.handler, Version: v, Priority: reg.priority})
						}
					}
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
