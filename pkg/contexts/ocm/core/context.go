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

	"github.com/open-component-model/ocm/pkg/contexts/config"
	cfgcpi "github.com/open-component-model/ocm/pkg/contexts/config/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const CONTEXT_TYPE = "ocm.context.gardener.cloud"

const CommonTransportFormat = ctf.CommonTransportFormatRepositoryType

type Context interface {
	datacontext.Context

	AttributesContext() datacontext.AttributesContext
	ConfigContext() config.Context
	CredentialsContext() credentials.Context
	OCIContext() oci.Context

	RepositoryTypes() RepositoryTypeScheme
	AccessMethods() AccessTypeScheme

	RepositorySpecHandlers() RepositorySpecHandlers
	MapUniformRepositorySpec(u *UniformRepositorySpec) (RepositorySpec, error)

	BlobHandlers() BlobHandlerRegistry
	BlobDigesters() BlobDigesterRegistry

	RepositoryForSpec(spec RepositorySpec, creds ...credentials.CredentialsSource) (Repository, error)
	RepositoryForConfig(data []byte, unmarshaler runtime.Unmarshaler, creds ...credentials.CredentialsSource) (Repository, error)
	AccessSpecForSpec(spec compdesc.AccessSpec) (AccessSpec, error)
	AccessSpecForConfig(data []byte, unmarshaler runtime.Unmarshaler) (AccessSpec, error)

	Encode(AccessSpec, runtime.Marshaler) ([]byte, error)

	GetAlias(name string) RepositorySpec
	SetAlias(name string, spec RepositorySpec)
}

////////////////////////////////////////////////////////////////////////////////

var key = reflect.TypeOf(_context{})

// DefaultContext is the default context initialized by init functions
var DefaultContext = Builder{}.New()

// ForContext returns the Context to use for context.Context.
// This is eiter an explicit context or the default context.
func ForContext(ctx context.Context) Context {
	return datacontext.ForContextByKey(ctx, key, DefaultContext).(Context)
}

////////////////////////////////////////////////////////////////////////////////

type _context struct {
	datacontext.Context
	updater cfgcpi.Updater

	sharedattributes datacontext.AttributesContext
	credctx          credentials.Context
	ocictx           oci.Context

	knownRepositoryTypes RepositoryTypeScheme
	knownAccessTypes     AccessTypeScheme

	specHandlers  RepositorySpecHandlers
	blobHandlers  BlobHandlerRegistry
	blobDigesters BlobDigesterRegistry
	aliases       map[string]RepositorySpec
}

var _ Context = &_context{}

func newContext(shared datacontext.AttributesContext, credctx credentials.Context, ocictx oci.Context, reposcheme RepositoryTypeScheme, accessscheme AccessTypeScheme, specHandlers RepositorySpecHandlers, blobHandlers BlobHandlerRegistry, blobDigesters BlobDigesterRegistry) Context {
	c := &_context{
		sharedattributes:     shared,
		updater:              cfgcpi.NewUpdate(ocictx.ConfigContext()),
		credctx:              credctx,
		ocictx:               ocictx,
		specHandlers:         specHandlers,
		blobHandlers:         blobHandlers,
		blobDigesters:        blobDigesters,
		knownAccessTypes:     accessscheme,
		knownRepositoryTypes: reposcheme,
		aliases:              map[string]RepositorySpec{},
	}
	c.Context = datacontext.NewContextBase(c, CONTEXT_TYPE, key, shared.GetAttributes())
	return c
}

func (c *_context) AttributesContext() datacontext.AttributesContext {
	return c.sharedattributes
}

func (c *_context) ConfigContext() config.Context {
	return c.updater.GetContext()
}

func (c *_context) CredentialsContext() credentials.Context {
	return c.credctx
}

func (c *_context) OCIContext() oci.Context {
	return c.ocictx
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

func (c *_context) BlobHandlers() BlobHandlerRegistry {
	return c.blobHandlers
}

func (c *_context) BlobDigesters() BlobDigesterRegistry {
	return c.blobDigesters
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

func (c *_context) AccessMethods() AccessTypeScheme {
	return c.knownAccessTypes
}

func (c *_context) AccessSpecForConfig(data []byte, unmarshaler runtime.Unmarshaler) (AccessSpec, error) {
	return c.knownAccessTypes.DecodeAccessSpec(data, unmarshaler)
}

func (c *_context) AccessSpecForSpec(spec compdesc.AccessSpec) (AccessSpec, error) {
	if spec == nil {
		return nil, nil
	}
	if n, ok := spec.(AccessSpec); ok {
		if g, ok := spec.(*GenericAccessSpec); ok {
			return g.Evaluate(c)
		}
		return n, nil
	}
	un, err := runtime.ToUnstructuredTypedObject(spec)
	if err != nil {
		return nil, err
	}

	raw, err := un.GetRaw()
	if err != nil {
		return nil, err
	}

	return c.knownAccessTypes.DecodeAccessSpec(raw, runtime.DefaultJSONEncoding)
}

func (c *_context) Encode(spec AccessSpec, marshaler runtime.Marshaler) ([]byte, error) {
	return c.knownAccessTypes.Encode(spec, marshaler)
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
