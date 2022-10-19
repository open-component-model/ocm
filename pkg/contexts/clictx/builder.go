// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package clictx

import (
	"context"
	"io"

	"github.com/mandelsoft/vfs/pkg/vfs"

	core2 "github.com/open-component-model/ocm/pkg/contexts/clictx/core"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
)

func WithContext(ctx context.Context) core2.Builder {
	return core2.Builder{}.WithContext(ctx)
}

func WithSharedAttributes(ctx datacontext.AttributesContext) core2.Builder {
	return core2.Builder{}.WithSharedAttributes(ctx)
}

func WithOCM(ctx ocm.Context) core2.Builder {
	return core2.Builder{}.WithOCM(ctx)
}

func WithFileSystem(fs vfs.FileSystem) core2.Builder {
	return core2.Builder{}.WithFileSystem(fs)
}

func WithOutput(w io.Writer) core2.Builder {
	return core2.Builder{}.WithOutput(w)
}

func WithErrorOutput(w io.Writer) core2.Builder {
	return core2.Builder{}.WithErrorOutput(w)
}

func WithInput(r io.Reader) core2.Builder {
	return core2.Builder{}.WithInput(r)
}

func New(mode ...datacontext.BuilderMode) core2.Context {
	return core2.Builder{}.New(mode...)
}
