// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package defaultconfigregistry

import (
	"slices"
	"sync"

	"github.com/open-component-model/ocm/pkg/contexts/config"
)

type DefaultConfigHandler func(cfg config.Context) error

type defaultConfigurationRegistry struct {
	lock sync.Mutex

	list []DefaultConfigHandler
}

func (r *defaultConfigurationRegistry) Register(h DefaultConfigHandler) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.list = append(r.list, h)
}

func (r *defaultConfigurationRegistry) Get() []DefaultConfigHandler {
	r.lock.Lock()
	defer r.lock.Unlock()

	return slices.Clone(r.list)
}

var defaultConfigRegistry = &defaultConfigurationRegistry{}

func RegisterDefaultConfigHandler(h DefaultConfigHandler) {
	defaultConfigRegistry.Register(h)
}

func Get() []DefaultConfigHandler {
	return defaultConfigRegistry.Get()
}
