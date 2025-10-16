package internal

import (
	"context"
	"io"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/utils/out"
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

func (b Builder) New(m ...datacontext.BuilderMode) Context {
	mode := datacontext.Mode(m...)
	ctx := b.getContext()

	if b.ocm == nil {
		var ok bool
		b.ocm, ok = ocm.DefinedForContext(ctx)
		if !ok && mode != datacontext.MODE_SHARED {
			b.ocm = ocm.New(mode)
		}
	}
	if b.shared == nil {
		b.shared = b.ocm.AttributesContext()
	}
	return datacontext.SetupContext(mode, newContext(b.shared, b.ocm, out.NewFor(b.out), b.filesystem, b.shared))
}
