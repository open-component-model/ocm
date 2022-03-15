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

package core

import (
	"context"
	"reflect"
	"sync"

	"github.com/gardener/ocm/pkg/config"
	cfgcpi "github.com/gardener/ocm/pkg/config/cpi"
	"github.com/gardener/ocm/pkg/credentials"
	"github.com/gardener/ocm/pkg/datacontext"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci"
	"github.com/gardener/ocm/pkg/ocm"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

const CONTEXT_TYPE = "ocm.cmd.context.gardener.cloud"

type Context interface {
	datacontext.Context

	AttributesContext() datacontext.AttributesContext

	ConfigContext() config.Context
	CredentialsContext() credentials.Context
	OCIContext() oci.Context
	OCMContext() ocm.Context

	FileSystem() vfs.FileSystem

	GetOCMRepository(name string) (ocm.Repository, error)
	GetOCIRepository(name string) (oci.Repository, error)

	AddOCIRepository(name string, spec oci.RepositorySpec) error
	AddOCMRepository(name string, spec ocm.RepositorySpec) error
}

var key = reflect.TypeOf(_context{})

// DefaultContext is the default context initialized by init functions
var DefaultContext = Builder{}.New()

// ForContext returns the Context to use for context.Context.
// This is eiter an explicit context or the default context.
// The returned context incorporates the given context.
func ForContext(ctx context.Context) Context {
	return datacontext.ForContextByKey(ctx, key, DefaultContext).(Context)
}

////////////////////////////////////////////////////////////////////////////////

type _context struct {
	lock sync.RWMutex
	datacontext.Context
	updater cfgcpi.Updater

	sharedAttributes datacontext.AttributesContext

	config      config.Context
	credentials credentials.Context
	oci         oci.Context
	ocm         ocm.Context

	filesystem vfs.FileSystem

	ocirepos map[string]oci.RepositorySpec
	ocmrepos map[string]ocm.RepositorySpec
}

var _ Context = &_context{}

func newContext(shared datacontext.AttributesContext, ocmctx ocm.Context, fs vfs.FileSystem) Context {
	if fs == nil {
		fs = osfs.New()
	}
	c := &_context{
		sharedAttributes: shared,
		ocm:              ocmctx,
		oci:              ocmctx.OCIContext(),
		credentials:      ocmctx.CredentialsContext(),
		config:           ocmctx.CredentialsContext().ConfigContext(),
		updater:          cfgcpi.NewUpdate(ocmctx.CredentialsContext().ConfigContext()),

		filesystem: fs,
		ocirepos:   map[string]oci.RepositorySpec{},
		ocmrepos:   map[string]ocm.RepositorySpec{},
	}
	c.Context = datacontext.NewContextBase(c, CONTEXT_TYPE, key, shared.GetAttributes())
	return c
}

func (c *_context) AttributesContext() datacontext.AttributesContext {
	return c.sharedAttributes
}

func (c *_context) ConfigContext() config.Context {
	return c.config
}

func (c *_context) CredentialsContext() credentials.Context {
	return c.credentials
}

func (c *_context) OCIContext() oci.Context {
	return c.oci
}

func (c *_context) OCMContext() ocm.Context {
	return c.ocm
}

func (c *_context) FileSystem() vfs.FileSystem {
	return c.filesystem
}

func (c *_context) AddOCIRepository(name string, spec oci.RepositorySpec) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.ocirepos[name] = spec
	return nil
}

func (c *_context) AddOCMRepository(name string, spec ocm.RepositorySpec) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.ocmrepos[name] = spec
	return nil
}

func (c *_context) GetOCIRepository(name string) (oci.Repository, error) {
	err := c.updater.Update(c)
	if err != nil {
		return nil, err
	}
	c.lock.RLock()
	defer c.lock.RUnlock()
	spec := c.ocirepos[name]

	if spec == nil {
		return nil, errors.ErrUnknown("oci repository", name)
	}
	return c.oci.RepositoryForSpec(spec)
}

func (c *_context) GetOCMRepository(name string) (ocm.Repository, error) {
	err := c.updater.Update(c)
	if err != nil {
		return nil, err
	}
	c.lock.RLock()
	defer c.lock.RUnlock()

	spec := c.ocmrepos[name]

	if spec == nil {
		return nil, errors.ErrUnknown("ocm repository", name)
	}
	return c.ocm.RepositoryForSpec(spec)
}
