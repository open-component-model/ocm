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
	"github.com/gardener/ocm/pkg/runtime"
)

type Context interface {
	datacontext.Context
	// With return the actual context for incorporating the given context.Context
	With(ctx context.Context) Context

	CredentialsContext() credentials.Context

	RepositoryTypes() RepositoryTypeScheme

	RepositoryForSpec(spec RepositorySpec, creds ...credentials.CredentialsSource) (Repository, error)
	RepositoryForConfig(data []byte, unmarshaler runtime.Unmarshaler, creds ...credentials.CredentialsSource) (Repository, error)
}

// DefaultContext is the default context initialized by init functions
var DefaultContext = NewContext(credentials.DefaultContext, DefaultRepositoryTypeScheme)

var key = reflect.TypeOf(_context{})

// ForContext returns the Context to use for context.Context.
// This is eiter an explicit context or the default context.
// The returned context incorporates the given context.
func ForContext(ctx context.Context) Context {
	c := ctx.Value(key)
	if c == nil {
		c = DefaultContext
	}
	return c.(Context).With(ctx)
}

////////////////////////////////////////////////////////////////////////////////

type _context struct {
	datacontext.DefaultContext
	data *_contextData // cache for correctly typed context data, repölaces
	// c.DefaultAccess().DataContext().(*_contextData)
}

var _ Context = &_context{}

func NewContext(ctx context.Context, reposcheme RepositoryTypeScheme) Context {
	return datacontext.NewContext(ctx, newDataContext(ctx, reposcheme)).(Context)
}

func (c *_context) With(ctx context.Context) Context {
	return c.DefaultAccess().With(ctx).(Context)
}

func (c *_context) CredentialsContext() credentials.Context {
	return c.data.credentials
}

func (c *_context) RepositoryTypes() RepositoryTypeScheme {
	return c.data.knownRepositoryTypes
}

func (c *_context) RepositorySpecForConfig(data []byte, unmarshaler runtime.Unmarshaler) (RepositorySpec, error) {
	return c.data.knownRepositoryTypes.DecodeRepositorySpec(data, unmarshaler)
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

////////////////////////////////////////////////////////////////////////////////

type _contextData struct {
	datacontext.AttributesContext

	credentials credentials.Context

	knownRepositoryTypes RepositoryTypeScheme
}

func newDataContext(ctx context.Context, reposcheme RepositoryTypeScheme) *_contextData {
	return &_contextData{
		AttributesContext:    datacontext.NewAttributes(nil),
		credentials:          credentials.ForContext(ctx),
		knownRepositoryTypes: reposcheme,
	}
}

func (c *_contextData) Wrap(defaultContext datacontext.DefaultContext) (datacontext.DefaultContext, interface{}) {
	return &_context{defaultContext, defaultContext.DefaultAccess().DataContext().(*_contextData)}, key
}
