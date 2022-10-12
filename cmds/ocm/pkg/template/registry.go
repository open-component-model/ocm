// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package template

import (
	"fmt"
	"strings"
	"sync"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/utils"
)

const KIND_TEMPLATER = "templater"

type TemplaterFactory func(system vfs.FileSystem) Templater

type Registry interface {
	Register(name string, fac TemplaterFactory, desc string)
	Create(name string, fs vfs.FileSystem) (Templater, error)
	Describe(name string) (string, error)
	KnownTypeNames() []string
}

type templaterInfo struct {
	templater   TemplaterFactory
	description string
}

type registry struct {
	lock       sync.RWMutex
	templaters map[string]templaterInfo
}

func NewRegistry() Registry {
	return &registry{
		templaters: map[string]templaterInfo{},
	}
}

func (r *registry) Register(name string, fac TemplaterFactory, desc string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.templaters[name] = templaterInfo{
		templater:   fac,
		description: desc,
	}
}

func (r *registry) Create(name string, fs vfs.FileSystem) (Templater, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	t, ok := r.templaters[name]
	if !ok {
		return nil, errors.ErrNotSupported(KIND_TEMPLATER, name)
	}
	return t.templater(fs), nil
}

func (r *registry) Describe(name string) (string, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	t, ok := r.templaters[name]
	if !ok {
		return "", errors.ErrNotSupported(KIND_TEMPLATER, name)
	}
	return t.description, nil
}

func (r *registry) KnownTypeNames() []string {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return utils.StringMapKeys(r.templaters)
}

func Usage(scheme Registry) string {
	s := `
There are several templaters that can be selected by the <code>--templater</code> option:
`
	for _, t := range scheme.KnownTypeNames() {
		desc, err := scheme.Describe(t)
		if err == nil {
			var title string
			idx := strings.Index(desc, "\n")
			if idx >= 0 {
				title = desc[:idx]
				desc = desc[idx+1:]
			}
			s = fmt.Sprintf("%s- <code>%s</code> %s\n\n%s", s, t, title, utils.IndentLines(desc, "  "))
			if !strings.HasSuffix(s, "\n") {
				s += "\n"
			}
		}
	}
	return s + "\n"
}

var _registry = NewRegistry()

func Register(name string, fac TemplaterFactory, desc string) {
	_registry.Register(name, fac, desc)
}

func DefaultRegistry() Registry {
	return _registry
}
