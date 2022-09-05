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

package download

import (
	"sync"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/out"
)

const ALL = "*"

type Handler interface {
	Download(ctx out.Context, racc cpi.ResourceAccess, path string, fs vfs.FileSystem) (bool, string, error)
}

type Registry interface {
	Register(typ string, hdlr Handler)
	Handler
	DownloadAsBlob(ctx out.Context, racc cpi.ResourceAccess, path string, fs vfs.FileSystem) (bool, string, error)
}

type registry struct {
	lock     sync.RWMutex
	handlers map[string][]Handler
}

func NewRegistry() Registry {
	return &registry{
		handlers: map[string][]Handler{},
	}
}

func (r *registry) Register(typ string, hdlr Handler) {
	r.lock.Lock()
	defer r.lock.Unlock()

	list := r.handlers[typ]
	list = append(list, hdlr)
	r.handlers[typ] = list
}

func (r *registry) getHandlers(typ string) []Handler {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.handlers[typ]
}

func (r *registry) Download(ctx out.Context, racc cpi.ResourceAccess, path string, fs vfs.FileSystem) (bool, string, error) {
	if ok, p, err := r.download(r.getHandlers(racc.Meta().GetType()), ctx, racc, path, fs); ok {
		return ok, p, err
	}
	return r.download(r.getHandlers(ALL), ctx, racc, path, fs)
}

func (r *registry) DownloadAsBlob(ctx out.Context, racc cpi.ResourceAccess, path string, fs vfs.FileSystem) (bool, string, error) {
	return r.download(r.getHandlers(ALL), ctx, racc, path, fs)
}

func (r *registry) download(list []Handler, ctx out.Context, racc cpi.ResourceAccess, path string, fs vfs.FileSystem) (bool, string, error) {
	errs := errors.ErrListf("download")
	for _, h := range list {
		ok, p, err := h.Download(ctx, racc, path, fs)
		if ok {
			return ok, p, err
		}
		errs.Add(err)
	}
	return false, "", errs.Result()
}

var DefaultRegistry = NewRegistry()

func Register(typ string, hdlr Handler) {
	DefaultRegistry.Register(typ, hdlr)
}

////////////////////////////////////////////////////////////////////////////////

const ATTR_DOWNLOADER_HANDLERS = "github.com/open-component-model/ocm/pkg/contexts/ocm/download"

func For(ctx datacontext.Context) Registry {
	if ctx == nil {
		return DefaultRegistry
	}
	return ctx.GetAttributes().GetAttribute(ATTR_DOWNLOADER_HANDLERS, DefaultRegistry).(Registry)
}

func SetFor(ctx datacontext.Context, registry Registry) {
	ctx.GetAttributes().SetAttribute(ATTR_DOWNLOADER_HANDLERS, registry)
}
