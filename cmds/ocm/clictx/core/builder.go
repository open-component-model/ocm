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
	"io"

	"github.com/gardener/ocm/cmds/ocm/pkg/output/out"
	"github.com/gardener/ocm/pkg/datacontext"
	"github.com/gardener/ocm/pkg/ocm"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

type Builder struct {
	ctx        context.Context
	shared     datacontext.AttributesContext
	ocm        ocm.Context
	out        out.Context
	filesystem vfs.FileSystem
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

func (b Builder) WithFileSystem(fs vfs.FileSystem) Builder {
	b.filesystem = fs
	return b
}

func (b Builder) WithSharedAttributes(ctx datacontext.AttributesContext) Builder {
	b.shared = ctx
	return b
}

func (b Builder) WithOCM(ctx ocm.Context) Builder {
	b.ocm = ctx
	return b
}

func (b Builder) WithOutput(w io.Writer) Builder {
	b.out = out.WithOutput(b.out, w)
	return b
}

func (b Builder) WithErrorOutput(w io.Writer) Builder {
	b.out = out.WithErrorOutput(b.out, w)
	return b
}

func (b Builder) WithInput(r io.Reader) Builder {
	b.out = out.WithInput(b.out, r)
	return b
}

func (b Builder) Bound() (Context, context.Context) {
	c := b.New()
	return c, context.WithValue(b.getContext(), key, c)
}

func (b Builder) New() Context {
	ctx := b.getContext()
	if b.ocm == nil {
		b.ocm = ocm.ForContext(ctx)
	}
	if b.shared == nil {
		b.shared = b.ocm.AttributesContext()
	}

	return newContext(b.shared, b.ocm, out.NewFor(b.out), b.filesystem)
}
