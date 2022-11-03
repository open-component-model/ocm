// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"encoding/json"
	"io"
	"sync"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/cache"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
	"github.com/open-component-model/ocm/pkg/errors"
)

type Plugin = *pluginImpl

type impl = cache.Plugin

// //nolint: errname // is no error.
type pluginImpl struct {
	lock sync.RWMutex
	impl
	config json.RawMessage
}

func NewPlugin(impl cache.Plugin, config json.RawMessage) Plugin {
	return &pluginImpl{
		impl:   impl,
		config: config,
	}
}

func (p *pluginImpl) SetConfig(config json.RawMessage) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.config = config
}

func (p *pluginImpl) Validate(spec []byte) (*ppi.AccessSpecInfo, error) {
	result, err := p.Exec(nil, nil, "accessmethod", "validate", string(spec))
	if err != nil {
		return nil, errors.Wrapf(err, "plugin %s", p.Name())
	}

	var info ppi.AccessSpecInfo
	err = json.Unmarshal(result, &info)
	if err != nil {
		return nil, errors.Wrapf(err, "plugin %s: cannot unmarshal access spec info", p.Name())
	}
	return &info, nil
}

func (p *pluginImpl) Exec(r io.Reader, w io.Writer, args ...string) ([]byte, error) {
	return cache.Exec(p.Path(), p.config, r, w, args...)
}
