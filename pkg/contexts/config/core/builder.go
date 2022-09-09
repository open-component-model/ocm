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

	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
)

type Builder struct {
	ctx        context.Context
	shared     datacontext.AttributesContext
	reposcheme ConfigTypeScheme
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

func (b Builder) WithConfigTypeScheme(scheme ConfigTypeScheme) Builder {
	b.reposcheme = scheme
	return b
}

func (b Builder) Bound() (Context, context.Context) {
	c := b.New()
	return c, context.WithValue(b.getContext(), key, c)
}

func (b Builder) New(m ...datacontext.BuilderMode) Context {
	mode := datacontext.Mode(m...)
	ctx := b.getContext()

	if b.shared == nil {
		if mode == datacontext.MODE_SHARED {
			b.shared = datacontext.ForContext(ctx)
		} else {
			b.shared = datacontext.New(nil)
		}
	}
	if b.reposcheme == nil {
		switch mode {
		case datacontext.MODE_INITIAL:
			b.reposcheme = NewConfigTypeScheme(nil)
		case datacontext.MODE_CONFIGURED:
			b.reposcheme = NewConfigTypeScheme(nil)
			b.reposcheme.AddKnownTypes(DefaultConfigTypeScheme)
		case datacontext.MODE_DEFAULTED:
			fallthrough
		case datacontext.MODE_SHARED:
			b.reposcheme = DefaultConfigTypeScheme
		}
	}
	return newContext(b.shared, b.reposcheme)
}
