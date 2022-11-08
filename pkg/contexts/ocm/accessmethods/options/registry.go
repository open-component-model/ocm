// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package options

import (
	"reflect"
	"sync"

	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
	errors "github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/utils"
)

const (
	KIND_OPTIONTYPE = "option type"
	KIND_OPTION     = "option"
)

type Creator func(name string, description string) flagsets.ConfigOptionType

type TypeInfo struct {
	Creator
	Description string
}

func (i TypeInfo) GetDescription() string {
	return i.Description
}

type Registry = *registry

var DefaultRegistry = New()

type registry struct {
	lock    sync.RWMutex
	types   map[string]TypeInfo
	options map[string]flagsets.ConfigOptionType
}

func New() Registry {
	return &registry{
		types:   map[string]TypeInfo{},
		options: map[string]flagsets.ConfigOptionType{},
	}
}

func (r *registry) RegisterOption(t flagsets.ConfigOptionType) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.options[t.GetName()] = t
}

func (r *registry) RegisterType(name string, c Creator, desc string) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.types[name] = TypeInfo{Creator: c, Description: desc}
}

func (r *registry) GetType(name string) *TypeInfo {
	r.lock.RLock()
	defer r.lock.RUnlock()
	if t, ok := r.types[name]; ok {
		return &t
	}
	return nil
}

func (r *registry) GetOption(name string) flagsets.ConfigOptionType {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.options[name]
}

func (r *registry) CreateOption(typ, name, desc string) (flagsets.ConfigOptionType, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	t, ok := r.types[typ]
	if !ok {
		return nil, errors.ErrUnknown(KIND_OPTIONTYPE, typ)
	}

	n := t.Creator(name, desc)
	o := r.options[name]
	if o != nil {
		if reflect.TypeOf(o) != reflect.TypeOf(n) {
			return nil, errors.ErrAlreadyExists(KIND_OPTION, name)
		}
		return o, nil
	}
	return n, nil
}

func (r *registry) Usage() string {
	r.lock.RLock()
	defer r.lock.RUnlock()

	tinfo := utils.FormatMap("", r.types)
	oinfo := utils.FormatMap("", r.options)

	return `
The following predifined options can be used:

` + oinfo + `

The following predefined option types are supported:

` + tinfo
}
