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

	"github.com/gardener/ocm/pkg/datacontext"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/runtime"
)

type Context interface {
	datacontext.Context
	// With returns the actual context for incorporating the given context.Context
	With(ctx context.Context) Context

	RepositoryTypes() RepositoryTypeScheme

	RepositorySpecForConfig(data []byte, unmarshaler runtime.Unmarshaler) (RepositorySpec, error)

	RepositoryForSpec(spec RepositorySpec, creds ...CredentialsSource) (Repository, error)
	RepositoryForConfig(data []byte, unmarshaler runtime.Unmarshaler, creds ...CredentialsSource) (Repository, error)

	CredentialsForSpec(spec CredentialsSpec, creds ...CredentialsSource) (Credentials, error)
	CredentialsForConfig(data []byte, unmarshaler runtime.Unmarshaler, cred ...CredentialsSource) (Credentials, error)

	GetCredentialsForConsumer(ConsumerIdentity) (Credentials, error)
	SetCredentialsForConsumer(identity ConsumerIdentity, creds CredentialsSource)

	SetAlias(name string, spec RepositorySpec, creds ...CredentialsSource) error
}

var key = reflect.TypeOf(_context{})

// DefaultContext is the default context initialized by init functions
var DefaultContext = NewContext(context.Background(), nil)

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
	return datacontext.NewContext(ctx, newDataContext(reposcheme)).(Context)
}

func (c *_context) With(ctx context.Context) Context {
	return c.DefaultAccess().With(ctx).(Context)
}

func (c *_context) RepositoryTypes() RepositoryTypeScheme {
	return c.data.knownRepositoryTypes
}

func (c *_context) RepositorySpecForConfig(data []byte, unmarshaler runtime.Unmarshaler) (RepositorySpec, error) {
	return c.data.knownRepositoryTypes.DecodeRepositorySpec(data, unmarshaler)
}

func (c *_context) RepositoryForSpec(spec RepositorySpec, creds ...CredentialsSource) (Repository, error) {
	cred, err := CredentialsChain(creds).Credentials(c)
	if err != nil {
		return nil, err
	}
	return spec.Repository(c, cred)
}

func (c *_context) RepositoryForConfig(data []byte, unmarshaler runtime.Unmarshaler, creds ...CredentialsSource) (Repository, error) {
	spec, err := c.data.knownRepositoryTypes.DecodeRepositorySpec(data, unmarshaler)
	if err != nil {
		return nil, err
	}
	return c.RepositoryForSpec(spec, creds...)
}

func (c *_context) CredentialsForSpec(spec CredentialsSpec, creds ...CredentialsSource) (Credentials, error) {
	repospec := spec.GetRepositorySpec(c)
	repo, err := c.RepositoryForSpec(repospec, creds...)
	if err != nil {
		return nil, err
	}
	return repo.LookupCredentials(spec.GetCredentialsName())

}

func (c *_context) CredentialsForConfig(data []byte, unmarshaler runtime.Unmarshaler, creds ...CredentialsSource) (Credentials, error) {
	spec := &GenericCredentialsSpec{}
	err := unmarshaler.Unmarshal(data, spec)
	if err != nil {
		return nil, err
	}
	return c.CredentialsForSpec(spec, creds...)
}

func (c *_context) GetCredentialsForConsumer(identity ConsumerIdentity) (Credentials, error) {
	consumer := c.data.consumers.Get(identity)
	if consumer == nil {
		return nil, ErrUnknownConsumer(identity.String())
	}
	return consumer.GetCredentials(c)
}

func (c *_context) SetCredentialsForConsumer(identity ConsumerIdentity, creds CredentialsSource) {
	c.data.consumers.Set(identity, creds)
}

func (c *_context) SetAlias(name string, spec RepositorySpec, creds ...CredentialsSource) error {
	t := c.data.knownRepositoryTypes.GetRepositoryType(AliasRepositoryType)
	if t == nil {
		return errors.ErrNotSupported("aliases")
	}
	if a, ok := t.(AliasRegistry); ok {
		return a.SetAlias(c, name, spec, CredentialsChain(creds))
	}
	return errors.ErrNotImplemented("interface", "AliasRegistry", reflect.TypeOf(t).String())
}

////////////////////////////////////////////////////////////////////////////////

type _contextData struct {
	datacontext.AttributesContext

	knownRepositoryTypes RepositoryTypeScheme
	consumers            *_consumers
}

var _ datacontext.DataContext = &_contextData{}

func newDataContext(reposcheme RepositoryTypeScheme) *_contextData {
	if reposcheme == nil {
		reposcheme = DefaultRepositoryTypeScheme
	}
	return &_contextData{
		AttributesContext:    datacontext.NewAttributes(nil),
		knownRepositoryTypes: reposcheme,
		consumers:            newConsumers(),
	}
}

func (c *_contextData) Wrap(defaultContext datacontext.DefaultContext) (datacontext.DefaultContext, interface{}) {
	return &_context{defaultContext, defaultContext.DefaultAccess().DataContext().(*_contextData)}, key
}
