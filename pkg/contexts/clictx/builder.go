// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package clictx

import (
	"context"
	"io"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/contexts/clictx/core"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
)

func WithContext(ctx context.Context) core.Builder {
	return core.Builder{}.WithContext(ctx)
}

func WithSharedAttributes(ctx datacontext.AttributesContext) core.Builder {
	return core.Builder{}.WithSharedAttributes(ctx)
}

func WithOCM(ctx ocm.Context) core.Builder {
	return core.Builder{}.WithOCM(ctx)
}

func WithFileSystem(fs vfs.FileSystem) core.Builder {
	return core.Builder{}.WithFileSystem(fs)
}

func WithOutput(w io.Writer) core.Builder {
	return core.Builder{}.WithOutput(w)
}

func WithErrorOutput(w io.Writer) core.Builder {
	return core.Builder{}.WithErrorOutput(w)
}

func WithInput(r io.Reader) core.Builder {
	return core.Builder{}.WithInput(r)
}

func New(mode ...datacontext.BuilderMode) core.Context {
	return core.Builder{}.New(mode...)
}
