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

	"github.com/gardener/ocm/pkg/credentials"
	"github.com/gardener/ocm/pkg/datacontext"
	"github.com/gardener/ocm/pkg/oci"
	"github.com/gardener/ocm/pkg/ocm/compdesc"
	"github.com/gardener/ocm/pkg/runtime"
)

type Context interface {
	datacontext.Context
	// With return the actual context for incorporating the given context.Context
	With(ctx context.Context) Context

	CredentialsContext() credentials.Context
	OCIContext() oci.Context

	RepositoryTypes() RepositoryTypeScheme
	AccessMethods() AccessTypeScheme

	RepositoryForSpec(spec RepositorySpec, creds ...credentials.CredentialsSource) (Repository, error)
	RepositoryForConfig(data []byte, unmarshaler runtime.Unmarshaler, creds ...credentials.CredentialsSource) (Repository, error)
	AccessSpecForSpec(spec compdesc.AccessSpec) (AccessSpec, error)
	AccessSpecForConfig(data []byte, unmarshaler runtime.Unmarshaler) (AccessSpec, error)
}

////////////////////////////////////////////////////////////////////////////////

var key = reflect.TypeOf(_context{})

// ForContextInternal returns the Context to use for context.Context.
func ForContextInternal(ctx context.Context) Context {
	c := ctx.Value(key)
	if c == nil {
		return nil
	}
	return c.(Context)
}

////////////////////////////////////////////////////////////////////////////////

type _context struct {
	datacontext.DefaultContext
	data *_contextData // cache for correctly typed context data, repölaces
	// c.DefaultAccess().DataContext().(*_contextData)
}

var _ Context = &_context{}

func NewContext(ctx context.Context, reposcheme RepositoryTypeScheme, accessscheme AccessTypeScheme) Context {
	return datacontext.NewContext(ctx, newDataContext(ctx, reposcheme, accessscheme)).(Context)
}

func (c *_context) With(ctx context.Context) Context {
	return c.DefaultAccess().With(ctx).(Context)
}

func (c *_context) CredentialsContext() credentials.Context {
	return c.data.ocictx.CredentialsContext()
}

func (c *_context) OCIContext() oci.Context {
	return c.data.ocictx
}

func (c *_context) RepositoryTypes() RepositoryTypeScheme {
	return c.data.knownRepositoryTypes
}

func (c *_context) RepositoryForSpec(spec RepositorySpec, creds ...credentials.CredentialsSource) (Repository, error) {
	cred, err := credentials.CredentialsChain(creds).Credentials(c.CredentialsContext())
	if err != nil {
		return nil, err
	}
	return spec.Repository(c, cred)
}

func (c *_context) RepositoryForConfig(data []byte, unmarshaler runtime.Unmarshaler, creds ...credentials.CredentialsSource) (Repository, error) {
	spec, err := c.data.knownRepositoryTypes.DecodeRepositorySpec(data, unmarshaler)
	if err != nil {
		return nil, err
	}
	return c.RepositoryForSpec(spec, creds...)
}

func (c *_context) AccessMethods() AccessTypeScheme {
	return c.data.knownAccessTypes
}

func (c *_context) AccessSpecForConfig(data []byte, unmarshaler runtime.Unmarshaler) (AccessSpec, error) {
	return c.data.knownAccessTypes.DecodeAccessSpec(data, unmarshaler)
}

func (c *_context) AccessSpecForSpec(spec compdesc.AccessSpec) (AccessSpec, error) {
	if spec == nil {
		return nil, nil
	}
	if n, ok := spec.(AccessSpec); ok {
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

	return c.data.knownAccessTypes.DecodeAccessSpec(raw, runtime.DefaultJSONEncoding)
}

////////////////////////////////////////////////////////////////////////////////

type _contextData struct {
	datacontext.AttributesContext

	ocictx oci.Context

	knownRepositoryTypes RepositoryTypeScheme
	knownAccessTypes     AccessTypeScheme
}

func newDataContext(ctx context.Context, reposcheme RepositoryTypeScheme, accessscheme AccessTypeScheme) *_contextData {
	if accessscheme == nil {
		accessscheme = DefaultAccessTypeScheme
	}
	if reposcheme == nil {
		reposcheme = DefaultRepositoryTypeScheme
	}
	return &_contextData{
		AttributesContext:    datacontext.NewAttributes(nil),
		ocictx:               oci.ForContext(ctx),
		knownAccessTypes:     accessscheme,
		knownRepositoryTypes: reposcheme,
	}
}

func (c *_contextData) Wrap(defaultContext datacontext.DefaultContext) (datacontext.DefaultContext, interface{}) {
	return &_context{defaultContext, defaultContext.DefaultAccess().DataContext().(*_contextData)}, key
}
