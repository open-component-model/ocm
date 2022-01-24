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
	"github.com/gardener/ocm/pkg/datacontext"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/runtime"
)

type Context interface {
	datacontext.Context

	AttributesContext() datacontext.AttributesContext

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
var DefaultContext = Builder{}.New()

// ForContext returns the Context to use for context.Context.
// This is eiter an explicit context or the default context.
func ForContext(ctx context.Context) Context {
	return datacontext.ForContextByKey(ctx, key, DefaultContext).(Context)
}

////////////////////////////////////////////////////////////////////////////////

type _context struct {
	datacontext.Context

	sharedattributes     datacontext.AttributesContext
	configctx            config.Context
	lock                 sync.RWMutex
	lastGeneration       int64
	knownRepositoryTypes RepositoryTypeScheme
	consumers            *_consumers
}

var _ Context = &_context{}

func newContext(shared datacontext.AttributesContext, configctx config.Context, reposcheme RepositoryTypeScheme) Context {
	c := &_context{
		sharedattributes:     shared,
		configctx:            configctx,
		knownRepositoryTypes: reposcheme,
		consumers:            newConsumers(),
	}
	c.Context = datacontext.NewContextBase(c, key, shared.GetAttributes())
	return c
}

func (c *_context) Update() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	gen, err := c.configctx.ApplyTo(c.lastGeneration, c)
	c.lastGeneration = gen
	return err
}

func (c *_context) AttributesContext() datacontext.AttributesContext {
	return c.sharedattributes
}

func (c *_context) RepositoryTypes() RepositoryTypeScheme {
	return c.knownRepositoryTypes
}

func (c *_context) RepositorySpecForConfig(data []byte, unmarshaler runtime.Unmarshaler) (RepositorySpec, error) {
	return c.knownRepositoryTypes.DecodeRepositorySpec(data, unmarshaler)
}

func (c *_context) RepositoryForSpec(spec RepositorySpec, creds ...CredentialsSource) (Repository, error) {
	cred, err := CredentialsChain(creds).Credentials(c)
	if err != nil {
		return nil, err
	}
	c.Update()
	return spec.Repository(c, cred)
}

func (c *_context) RepositoryForConfig(data []byte, unmarshaler runtime.Unmarshaler, creds ...CredentialsSource) (Repository, error) {
	spec, err := c.knownRepositoryTypes.DecodeRepositorySpec(data, unmarshaler)
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
	c.Update()
	consumer := c.consumers.Get(identity)
	if consumer == nil {
		return nil, ErrUnknownConsumer(identity.String())
	}
	return consumer.GetCredentials(c)
}

func (c *_context) SetCredentialsForConsumer(identity ConsumerIdentity, creds CredentialsSource) {
	c.Update()
	c.consumers.Set(identity, creds)
}

func (c *_context) SetAlias(name string, spec RepositorySpec, creds ...CredentialsSource) error {
	c.Update()
	t := c.knownRepositoryTypes.GetRepositoryType(AliasRepositoryType)
	if t == nil {
		return errors.ErrNotSupported("aliases")
	}
	if a, ok := t.(AliasRegistry); ok {
		return a.SetAlias(c, name, spec, CredentialsChain(creds))
	}
	return errors.ErrNotImplemented("interface", "AliasRegistry", reflect.TypeOf(t).String())
}
