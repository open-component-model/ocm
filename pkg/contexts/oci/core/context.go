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
	"strings"

	"github.com/mandelsoft/logging"

	"github.com/open-component-model/ocm/pkg/contexts/config"
	cfgcpi "github.com/open-component-model/ocm/pkg/contexts/config/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const CONTEXT_TYPE = "oci.context.gardener.cloud"

const CommonTransportFormat = "CommonTransportFormat"

type Context interface {
	datacontext.Context

	AttributesContext() datacontext.AttributesContext
	ConfigContext() config.Context
	CredentialsContext() credentials.Context

	RepositorySpecHandlers() RepositorySpecHandlers
	MapUniformRepositorySpec(u *UniformRepositorySpec) (RepositorySpec, error)

	RepositoryTypes() RepositoryTypeScheme

	RepositoryForSpec(spec RepositorySpec, creds ...credentials.CredentialsSource) (Repository, error)
	RepositoryForConfig(data []byte, unmarshaler runtime.Unmarshaler, creds ...credentials.CredentialsSource) (Repository, error)

	GetAlias(name string) RepositorySpec
	SetAlias(name string, spec RepositorySpec)
}

var key = reflect.TypeOf(_context{})

// DefaultContext is the default context initialized by init functions.
var DefaultContext = Builder{}.New(datacontext.MODE_SHARED)

// ForContext returns the Context to use for context.Context.
// This is either an explicit context or the default context.
func ForContext(ctx context.Context) Context {
	c, _ := datacontext.ForContextByKey(ctx, key, DefaultContext)
	return c.(Context)
}

func DefinedForContext(ctx context.Context) (Context, bool) {
	c, ok := datacontext.ForContextByKey(ctx, key, DefaultContext)
	if c != nil {
		return c.(Context), ok
	}
	return nil, ok
}

////////////////////////////////////////////////////////////////////////////////

type _context struct {
	datacontext.Context
	updater cfgcpi.Updater

	sharedattributes datacontext.AttributesContext
	credentials      credentials.Context

	knownRepositoryTypes RepositoryTypeScheme
	specHandlers         RepositorySpecHandlers
	aliases              map[string]RepositorySpec
}

func newContext(creds credentials.Context, reposcheme RepositoryTypeScheme, specHandlers RepositorySpecHandlers, logger logging.Context) Context {
	c := &_context{
		sharedattributes:     creds.AttributesContext(),
		updater:              cfgcpi.NewUpdate(creds.ConfigContext()),
		credentials:          creds,
		knownRepositoryTypes: reposcheme,
		specHandlers:         specHandlers,
		aliases:              map[string]RepositorySpec{},
	}
	c.Context = datacontext.NewContextBase(c, CONTEXT_TYPE, key, creds.ConfigContext().GetAttributes(), logger)
	return c
}

var _ Context = &_context{}

func (c *_context) AttributesContext() datacontext.AttributesContext {
	return c.sharedattributes
}

func (c *_context) ConfigContext() config.Context {
	return c.updater.GetContext()
}

func (c *_context) CredentialsContext() credentials.Context {
	return c.credentials
}

func (c *_context) RepositoryTypes() RepositoryTypeScheme {
	return c.knownRepositoryTypes
}

func (c *_context) RepositorySpecHandlers() RepositorySpecHandlers {
	return c.specHandlers
}

func (c *_context) MapUniformRepositorySpec(u *UniformRepositorySpec) (RepositorySpec, error) {
	return c.specHandlers.MapUniformRepositorySpec(c, u)
}

func (c *_context) RepositorySpecForConfig(data []byte, unmarshaler runtime.Unmarshaler) (RepositorySpec, error) {
	return c.knownRepositoryTypes.DecodeRepositorySpec(data, unmarshaler)
}

func (c *_context) RepositoryForSpec(spec RepositorySpec, creds ...credentials.CredentialsSource) (Repository, error) {
	cred, err := credentials.CredentialsChain(creds).Credentials(c.CredentialsContext())
	if err != nil {
		return nil, err
	}
	return spec.Repository(c, cred)
}

func (c *_context) RepositoryForConfig(data []byte, unmarshaler runtime.Unmarshaler, creds ...credentials.CredentialsSource) (Repository, error) {
	spec, err := c.knownRepositoryTypes.DecodeRepositorySpec(data, unmarshaler)
	if err != nil {
		return nil, err
	}
	return c.RepositoryForSpec(spec, creds...)
}

func (c *_context) GetAlias(name string) RepositorySpec {
	err := c.updater.Update(c)
	if err != nil {
		return nil
	}
	c.updater.RLock()
	defer c.updater.RUnlock()
	spec := c.aliases[name]
	if spec == nil && strings.HasSuffix(name, ".alias") {
		spec = c.aliases[name[:len(name)-6]]
	}
	return spec
}

func (c *_context) SetAlias(name string, spec RepositorySpec) {
	c.updater.Lock()
	defer c.updater.Unlock()
	c.aliases[name] = spec
}
