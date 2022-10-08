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
	"fmt"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/logging"
)

type Builder struct {
	ctx           context.Context
	credentials   credentials.Context
	oci           oci.Context
	reposcheme    RepositoryTypeScheme
	accessscheme  AccessTypeScheme
	spechandlers  RepositorySpecHandlers
	blobhandlers  BlobHandlerRegistry
	blobdigesters BlobDigesterRegistry
}

func (b *Builder) getContext() context.Context {
	if b.ctx == nil {
		return context.Background()
	}
	return b.ctx
}

func (b Builder) WithContext(ctx context.Context) Builder {
	b.ctx = ctx
	return b
}

func (b Builder) WithCredentials(ctx credentials.Context) Builder {
	b.credentials = ctx
	return b
}

func (b Builder) WithOCIRepositories(ctx oci.Context) Builder {
	b.oci = ctx
	return b
}

func (b Builder) WithRepositoyTypeScheme(scheme RepositoryTypeScheme) Builder {
	b.reposcheme = scheme
	return b
}

func (b Builder) WithAccessTypeScheme(scheme AccessTypeScheme) Builder {
	b.accessscheme = scheme
	return b
}

func (b Builder) WithRepositorySpecHandlers(reg RepositorySpecHandlers) Builder {
	b.spechandlers = reg
	return b
}

func (b Builder) WithBlobHandlers(reg BlobHandlerRegistry) Builder {
	b.blobhandlers = reg
	return b
}

func (b Builder) WithBlobDigesters(reg BlobDigesterRegistry) Builder {
	b.blobdigesters = reg
	return b
}

func (b Builder) Bound() (Context, context.Context) {
	c := b.New()
	return c, context.WithValue(b.getContext(), key, c)
}

func (b Builder) New(m ...datacontext.BuilderMode) Context {
	mode := datacontext.Mode(m...)
	ctx := b.getContext()

	if b.oci == nil {
		if b.credentials != nil {
			b.oci = oci.WithCredentials(b.credentials).New(mode)
		} else {
			var ok bool
			b.oci, ok = oci.DefinedForContext(ctx)
			if !ok && mode != datacontext.MODE_SHARED {
				b.oci = oci.New(mode)
			}
		}
	}
	if b.credentials == nil {
		b.credentials = b.oci.CredentialsContext()
	}
	if b.reposcheme == nil {
		switch mode {
		case datacontext.MODE_INITIAL:
			b.reposcheme = NewRepositoryTypeScheme(nil)
		case datacontext.MODE_CONFIGURED:
			b.reposcheme = NewRepositoryTypeScheme(nil)
			b.reposcheme.AddKnownTypes(DefaultRepositoryTypeScheme)
		case datacontext.MODE_DEFAULTED:
			fallthrough
		case datacontext.MODE_SHARED:
			b.reposcheme = DefaultRepositoryTypeScheme
		}
	}
	if b.accessscheme == nil {
		switch mode {
		case datacontext.MODE_INITIAL:
			b.accessscheme = NewAccessTypeScheme()
		case datacontext.MODE_CONFIGURED:
			b.accessscheme = NewAccessTypeScheme()
			b.accessscheme.AddKnownTypes(DefaultAccessTypeScheme)
		case datacontext.MODE_DEFAULTED:
			fallthrough
		case datacontext.MODE_SHARED:
			b.accessscheme = DefaultAccessTypeScheme
		}
	}
	if b.spechandlers == nil {
		switch mode {
		case datacontext.MODE_INITIAL:
			b.spechandlers = NewRepositorySpecHandlers()
		case datacontext.MODE_CONFIGURED:
			b.spechandlers = DefaultRepositorySpecHandlers.Copy()
		case datacontext.MODE_DEFAULTED:
			fallthrough
		case datacontext.MODE_SHARED:
			b.spechandlers = DefaultRepositorySpecHandlers
		}
	}
	if b.blobhandlers == nil {
		switch mode {
		case datacontext.MODE_INITIAL:
			b.blobhandlers = NewBlobHandlerRegistry()
		case datacontext.MODE_CONFIGURED:
			b.blobhandlers = DefaultBlobHandlerRegistry.Copy()
		case datacontext.MODE_DEFAULTED:
			fallthrough
		case datacontext.MODE_SHARED:
			b.blobhandlers = DefaultBlobHandlerRegistry
		}
	}
	if b.blobdigesters == nil {
		switch mode {
		case datacontext.MODE_INITIAL:
			b.blobdigesters = NewBlobDigesterRegistry()
		case datacontext.MODE_CONFIGURED:
			b.blobdigesters = DefaultBlobDigesterRegistry.Copy()
		case datacontext.MODE_DEFAULTED:
			fallthrough
		case datacontext.MODE_SHARED:
			b.blobdigesters = DefaultBlobDigesterRegistry
		}
	}

	if ociimpl != nil {
		def, err := ociimpl(b.oci)
		if err != nil {
			panic(fmt.Sprintf("cannot create oci default decoder: %s", err))
		}
		reposcheme := NewRepositoryTypeScheme(def)
		reposcheme.AddKnownTypes(b.reposcheme) // TODO: implement delegation
		b.reposcheme = reposcheme
	}
	logger := logging.NewDefaultContext()
	return newContext(b.credentials, b.oci, b.reposcheme, b.accessscheme, b.spechandlers, b.blobhandlers, b.blobdigesters, logger)
}
