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

	"github.com/open-component-model/ocm/pkg/config"
	"github.com/open-component-model/ocm/pkg/datacontext"
)

type Builder struct {
	ctx        context.Context
	shared     datacontext.AttributesContext
	config     config.Context
	reposcheme RepositoryTypeScheme
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

func (b Builder) WithConfig(ctx config.Context) Builder {
	b.config = ctx
	return b
}

func (b Builder) WithRepositoyTypeScheme(scheme RepositoryTypeScheme) Builder {
	b.reposcheme = scheme
	return b
}

func (b Builder) Bound() (Context, context.Context) {
	c := b.New()
	return c, context.WithValue(b.getContext(), key, c)
}

func (b Builder) New() Context {
	ctx := b.getContext()
	if b.config == nil {
		b.config = config.ForContext(ctx)
	}
	if b.shared == nil {
		b.shared = b.config.AttributesContext()
	}
	if b.reposcheme == nil {
		b.reposcheme = DefaultRepositoryTypeScheme
	}
	return newContext(b.shared, b.config, b.reposcheme)

}
