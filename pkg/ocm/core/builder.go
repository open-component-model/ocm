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

	"github.com/gardener/ocm/pkg/credentials"
	"github.com/gardener/ocm/pkg/datacontext"
	"github.com/gardener/ocm/pkg/oci"
)

type Builder struct {
	ctx          context.Context
	shared       datacontext.AttributesContext
	credentials  credentials.Context
	oci          oci.Context
	reposcheme   RepositoryTypeScheme
	accessscheme AccessTypeScheme
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

func (b Builder) WithSharedAttributes(ctx datacontext.AttributesContext) Builder {
	b.shared = ctx
	return b
}

func (b Builder) WithCredentials(ctx credentials.Context) Builder {
	b.shared = ctx
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

func (b Builder) Bound() (Context, context.Context) {
	c := b.New()
	return c, context.WithValue(b.getContext(), key, c)
}

func (b Builder) New() Context {
	ctx := b.getContext()

	if b.oci == nil {
		b.oci = oci.ForContext(ctx)
	}
	if b.credentials == nil {
		b.credentials = b.oci.CredentialsContext()
	}
	if b.shared == nil {
		b.shared = b.credentials.AttributesContext()
	}
	if b.reposcheme == nil {
		b.reposcheme = DefaultRepositoryTypeScheme
	}
	if b.accessscheme == nil {
		b.accessscheme = DefaultAccessTypeScheme
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
	return newContext(b.shared, b.credentials, b.oci, b.reposcheme, b.accessscheme)

}
