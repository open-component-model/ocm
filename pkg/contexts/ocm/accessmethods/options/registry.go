// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package options

import (
	"reflect"
	"sync"

	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
	errors "github.com/open-component-model/ocm/pkg/errors"
)

const (
	KIND_OPTIONTYPE = "option type"
	KIND_OPTION     = "option"
)

type Creator func(name string, description string) flagsets.ConfigOptionType

type Registry = *registry

var DefaultRegistry = New()

type registry struct {
	lock    sync.RWMutex
	types   map[string]Creator
	options map[string]flagsets.ConfigOptionType
}

func New() Registry {
	return &registry{
		types:   map[string]Creator{},
		options: map[string]flagsets.ConfigOptionType{},
	}
}

func (r *registry) RegisterOption(t flagsets.ConfigOptionType) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.options[t.Name()] = t
}

func (r *registry) RegisterType(name string, c Creator) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.types[name] = c
}

func (r *registry) GetType(name string) Creator {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.types[name]
}

func (r *registry) GetOption(name string) flagsets.ConfigOptionType {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.options[name]
}

func (r *registry) CreateOption(typ, name, desc string) (flagsets.ConfigOptionType, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	t := r.types[typ]
	if t == nil {
		return nil, errors.ErrUnknown(KIND_OPTIONTYPE, typ)
	}

	n := t(name, desc)
	o := r.options[name]
	if o != nil {
		if reflect.TypeOf(o) != reflect.TypeOf(n) {
			return nil, errors.ErrAlreadyExists(KIND_OPTION, name)
		}
		return o, nil
	}
	return n, nil
}
