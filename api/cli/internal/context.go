package internal

import (
	"context"
	"io"
	"reflect"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/config"
	cfgcpi "ocm.software/ocm/api/config/cpi"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	"ocm.software/ocm/api/oci"
	ctfoci "ocm.software/ocm/api/oci/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm"
	ctfocm "ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/out"
)

const CONTEXT_TYPE = "ocm.cmd" + datacontext.OCM_CONTEXT_SUFFIX

type OCI interface {
	Context() oci.Context
	OpenCTF(path string) (oci.Repository, error)
}

type OCM interface {
	Context() ocm.Context
	OpenCTF(path string) (ocm.Repository, error)
}

type FileSystem struct {
	vfs.FileSystem
}

var _ osfs.OsFsCheck = (*FileSystem)(nil)

func (f *FileSystem) IsOsFileSystem() bool {
	return osfs.IsOsFileSystem(f.FileSystem)
}

func (f *FileSystem) ApplyOption(options accessio.Options) error {
	options.SetPathFileSystem(f.FileSystem)
	return nil
}

type ContextProvider interface {
	CLIContext() Context
}

type Context interface {
	datacontext.Context
	ContextProvider
	datacontext.ContextProvider
	config.ContextProvider
	credentials.ContextProvider
	oci.ContextProvider
	ocm.ContextProvider

	FileSystem() *FileSystem

	OCI() OCI
	OCM() OCM

	ApplyOption(options accessio.Options) error

	out.Context
	WithStdIO(r io.Reader, o io.Writer, e io.Writer) Context
}

var key = reflect.TypeOf(_context{})

// DefaultContext is the default context initialized by init functions.
var DefaultContext = Builder{}.New(datacontext.MODE_SHARED)

// ForContext returns the Context to use for context.Context.
// This is either an explicit context or the default context.
// The returned context incorporates the given context.
func ForContext(ctx context.Context) Context {
	c, _ := datacontext.ForContextByKey(ctx, key, DefaultContext)
	return c.(Context)
}

func DefinedForContext(ctx context.Context) (Context, bool) {
	c, ok := datacontext.ForContextByKey(ctx, key, DefaultContext)
	if c != nil {
		return c.(Context), ok
	}
	return nil, ok
}

////////////////////////////////////////////////////////////////////////////////

type _InternalContext = datacontext.InternalContext

type _context struct {
	_InternalContext
	updater cfgcpi.Updater

	sharedAttributes datacontext.AttributesContext

	credentials credentials.Context
	oci         *_oci
	ocm         *_ocm

	out out.Context
}

var (
	_ Context                          = (*_context)(nil)
	_ datacontext.ViewCreator[Context] = (*_context)(nil)
)

// gcWrapper is used as garbage collectable
// wrapper for a context implementation
// to establish a runtime finalizer.
type gcWrapper struct {
	datacontext.GCWrapper
	*_context
}

func newView(c *_context, ref ...bool) Context {
	if general.Optional(ref...) {
		return datacontext.FinalizedContext[gcWrapper](c)
	}
	return c
}

func (w *gcWrapper) SetContext(c *_context) {
	w._context = c
}

func newContext(shared datacontext.AttributesContext, ocmctx ocm.Context, outctx out.Context, fs vfs.FileSystem, delegates datacontext.Delegates) Context {
	if outctx == nil {
		outctx = out.New()
	}
	if shared == nil {
		shared = ocmctx.AttributesContext()
	}
	c := &_context{
		sharedAttributes: datacontext.PersistentContextRef(shared),
		credentials:      datacontext.PersistentContextRef(ocmctx.CredentialsContext()),
		out:              outctx,
	}
	c._InternalContext = datacontext.NewContextBase(c, CONTEXT_TYPE, key, ocmctx.GetAttributes(), delegates)
	c.updater = cfgcpi.NewUpdater(datacontext.PersistentContextRef(ocmctx.CredentialsContext().ConfigContext()), c)
	ocmctx = datacontext.PersistentContextRef(ocmctx)
	c.oci = newOCI(c, ocmctx)
	c.ocm = newOCM(c, ocmctx)
	if fs != nil {
		vfsattr.Set(c.AttributesContext(), fs)
	}
	return newView(c, true)
}

func (c *_context) CreateView() Context {
	return newView(c, true)
}

func (c *_context) CLIContext() Context {
	return newView(c)
}

func (c *_context) Update() error {
	return c.updater.Update()
}

func (c *_context) AttributesContext() datacontext.AttributesContext {
	return c.sharedAttributes
}

func (c *_context) ConfigContext() config.Context {
	return c.updater.GetContext()
}

func (c *_context) CredentialsContext() credentials.Context {
	return c.credentials
}

func (c *_context) OCIContext() oci.Context {
	return c.oci.Context()
}

func (c *_context) OCMContext() ocm.Context {
	return c.ocm.Context()
}

func (c *_context) FileSystem() *FileSystem {
	return &FileSystem{vfsattr.Get(c.CLIContext())}
}

func (c *_context) OCI() OCI {
	return c.oci
}

func (c *_context) OCM() OCM {
	return c.ocm
}

func (c *_context) ApplyOption(options accessio.Options) error {
	options.SetPathFileSystem(c.FileSystem())
	return nil
}

func (c *_context) StdOut() io.Writer {
	return c.out.StdOut()
}

func (c *_context) StdErr() io.Writer {
	return c.out.StdErr()
}

func (c *_context) StdIn() io.Reader {
	return c.out.StdIn()
}

func (c *_context) WithStdIO(r io.Reader, o io.Writer, e io.Writer) Context {
	return &_view{
		Context: c.CLIContext(),
		out:     out.NewFor(out.WithStdIO(c.out, r, o, e)),
	}
}

////////////////////////////////////////////////////////////////////////////////

type _view struct {
	Context
	out out.Context
}

func (c *_view) StdOut() io.Writer {
	return c.out.StdOut()
}

func (c *_view) StdErr() io.Writer {
	return c.out.StdErr()
}

func (c *_view) StdIn() io.Reader {
	return c.out.StdIn()
}

func (c *_view) WithStdIO(r io.Reader, o io.Writer, e io.Writer) Context {
	return &_view{
		Context: c.CLIContext(),
		out:     out.NewFor(out.WithStdIO(c.out, r, o, e)),
	}
}

////////////////////////////////////////////////////////////////////////////////
// the coding for _oci and _ocm is identical except the package used:
// _oci uses oci and ctfoci
// _ocm uses ocm and ctfocm

type _oci struct {
	cli   *_context
	ctx   oci.Context
	repos map[string]oci.RepositorySpec
}

func newOCI(ctx *_context, ocmctx ocm.Context) *_oci {
	return &_oci{
		cli:   ctx,
		ctx:   ocmctx.OCIContext(),
		repos: map[string]oci.RepositorySpec{},
	}
}

func (c *_oci) Context() oci.Context {
	return c.ctx
}

func (c *_oci) OpenCTF(path string) (oci.Repository, error) {
	ok, err := vfs.Exists(c.cli.FileSystem(), path)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.ErrNotFound("file", path)
	}
	return ctfoci.Open(c.ctx, accessobj.ACC_WRITABLE, path, 0, accessio.PathFileSystem(c.cli.FileSystem()))
}

////////////////////////////////////////////////////////////////////////////////

type _ocm struct {
	cli   *_context
	ctx   ocm.Context
	repos map[string]ocm.RepositorySpec
}

func newOCM(ctx *_context, ocmctx ocm.Context) *_ocm {
	return &_ocm{
		cli:   ctx,
		ctx:   ocmctx,
		repos: map[string]ocm.RepositorySpec{},
	}
}

func (c *_ocm) Context() ocm.Context {
	return c.ctx
}

func (c *_ocm) OpenCTF(path string) (ocm.Repository, error) {
	ok, err := vfs.Exists(c.cli.FileSystem(), path)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.ErrNotFound("file", path)
	}
	return ctfocm.Open(c.ctx, accessobj.ACC_WRITABLE, path, 0, c.cli.FileSystem())
}
