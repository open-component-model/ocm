// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package defaultconfigregistry

import (
	"strings"
	"sync"

	"github.com/open-component-model/ocm/pkg/contexts/config"
	"github.com/open-component-model/ocm/pkg/listformat"
)

type DefaultConfigHandler func(cfg config.Context) error

type defaultConfigurationRegistry struct {
	lock sync.Mutex

	list []entry
}

type entry struct {
	desc    string
	handler DefaultConfigHandler
}

func (r *defaultConfigurationRegistry) Register(h DefaultConfigHandler, desc string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.list = append(r.list, entry{desc, h})
}

func (r *defaultConfigurationRegistry) Get() []DefaultConfigHandler {
	r.lock.Lock()
	defer r.lock.Unlock()

	var result []DefaultConfigHandler
	for _, h := range r.list {
		result = append(result, h.handler)
	}
	return result
}

var defaultConfigRegistry = &defaultConfigurationRegistry{}

func RegisterDefaultConfigHandler(h DefaultConfigHandler, desc string) {
	defaultConfigRegistry.Register(h, desc)
}

func Get() []DefaultConfigHandler {
	return defaultConfigRegistry.Get()
}

func Description() string {
	var result []string

	for _, h := range defaultConfigRegistry.list {
		if h.desc != "" {
			result = append(result, strings.TrimSpace(h.desc))
		}
	}
	return listformat.FormatDescriptionList("", result...)
}
