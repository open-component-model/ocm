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

package inputs

import (
	"sync"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/cmds/ocm/clictx"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type InputHandler interface {
	RequireFilePath() bool
	Validate(fldPath *field.Path, ctx clictx.Context, input *BlobInput) field.ErrorList
	CreateBlob(fs vfs.FileSystem, inputPath string, input *BlobInput) (accessio.TemporaryBlobAccess, string, error)
}

type Registry interface {
	Register(name string, handler InputHandler)
	Get(name string) InputHandler
}

type registry struct {
	lock     sync.RWMutex
	handlers map[string]InputHandler
}

func NewRegistry() Registry {
	return &registry{
		handlers: map[string]InputHandler{},
	}
}

func (r *registry) Register(name string, handler InputHandler) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.handlers[name] = handler
}

func (r *registry) Get(name string) InputHandler {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.handlers[name]
}

var Default = NewRegistry()

func Register(name string, handler InputHandler) {
	Default.Register(name, handler)
}

func Get(name string) InputHandler {
	return Default.Get(name)
}
