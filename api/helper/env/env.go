package env

import (
	"bytes"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/DataDog/gostackparse"
	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/goutils/exception"
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/pkgutils"
	"github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/vfs/pkg/composefs"
	"github.com/mandelsoft/vfs/pkg/layerfs"
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/readonlyfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/config"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	"ocm.software/ocm/api/oci"
	ocm "ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/accessio"
)

////////////////////////////////////////////////////////////////////////////////

// Option is he option interface for env creations.
// An Option just provides an OptionHandler
// which is used by the env creation to get info
// (like getting the ocm context)
// or to do something (like fs mounting).
type Option interface {
	OptionHandler() OptionHandler
}

// OptionHandler is the interface for the option actions.
// This indirection (Option -> OptionHandler) is introduced
// to enable objects to be usable as env option
// (for example Environment) without the need to pollute its
// interface with the effective option methods defined by
// OptionHandler. This would make no sense, because an option
// typically does nothing but for a selected set of methods
// according to its intended functionality. Nevertheless,
// is has to implement all the interface methods.
type OptionHandler interface {
	OCMContext() ocm.Context
	GetFilesystem() vfs.FileSystem
	GetFailHandler() FailHandler
	GetEnvironment() *Environment

	// actions on environment of properties

	// Mount mounts a new filesystem to the actual env filesystem.
	Mount(fs *composefs.ComposedFileSystem) error

	// Propagate is called on final environment.
	Propagate(e *Environment)
}

type DefaultOptionHandler struct{}

var _ OptionHandler = (*DefaultOptionHandler)(nil)

func (o DefaultOptionHandler) Propagate(e *Environment) {
}

func (o DefaultOptionHandler) OCMContext() ocm.Context {
	return nil
}

func (o DefaultOptionHandler) GetFilesystem() vfs.FileSystem {
	return nil
}

func (o DefaultOptionHandler) GetFailHandler() FailHandler {
	return nil
}

func (o DefaultOptionHandler) GetEnvironment() *Environment {
	return nil
}

func (DefaultOptionHandler) Mount(*composefs.ComposedFileSystem) error {
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type FailHandler func(msg string, callerSkip ...int)

func (f FailHandler) OptionHandler() OptionHandler {
	return f
}

func (f FailHandler) OCMContext() ocm.Context {
	return nil
}

func (f FailHandler) GetFailHandler() FailHandler {
	return f
}

func (FailHandler) GetFilesystem() vfs.FileSystem {
	return nil
}

func (FailHandler) GetEnvironment() *Environment {
	return nil
}

func (FailHandler) Mount(*composefs.ComposedFileSystem) error {
	return nil
}

func (FailHandler) Propagate(e *Environment) {
}

////////////////////////////////////////////////////////////////////////////////

type fsOpt struct {
	DefaultOptionHandler
	path string
	fs   vfs.FileSystem
}

func FileSystem(fs vfs.FileSystem, path ...string) Option {
	return fsOpt{
		path: utils.Optional(path...),
		fs:   fs,
	}
}

func (o fsOpt) OptionHandler() OptionHandler {
	return o
}

func (o fsOpt) GetFilesystem() vfs.FileSystem {
	if o.path == "" {
		return o.fs
	}
	return nil
}

func (o fsOpt) Mount(cfs *composefs.ComposedFileSystem) error {
	if o.path == "" {
		return nil
	}
	return cfs.Mount(o.path, o.fs)
}

////////////////////////////////////////////////////////////////////////////////

type ctxOpt struct {
	DefaultOptionHandler
	ctx ocm.Context
}

func OCMContext(ctx ocm.Context) Option {
	return ctxOpt{
		ctx: ctx,
	}
}

func (o ctxOpt) OptionHandler() OptionHandler {
	return o
}

func (o ctxOpt) OCMContext() ocm.Context {
	return o.ctx
}

////////////////////////////////////////////////////////////////////////////////

type propOpt struct {
	DefaultOptionHandler
}

func UseAsContextFileSystem() Option {
	return propOpt{}
}

func (o propOpt) OptionHandler() OptionHandler {
	return o
}

func (o ctxOpt) Propagate(e *Environment) {
	vfsattr.Set(e.OCMContext().AttributesContext(), e.FileSystem())
}

////////////////////////////////////////////////////////////////////////////////

type tdOpt struct {
	DefaultOptionHandler
	path       string
	source     string
	modifiable bool
}

func testData(modifiable bool, paths ...string) tdOpt {
	path := "/testdata"
	source := "testdata"

	switch len(paths) {
	case 0:
	case 1:
		source = paths[0]
	case 2:
		source = paths[0]
		path = paths[1]
	default:
		panic("invalid number of arguments")
	}
	return tdOpt{
		path:       path,
		source:     source,
		modifiable: modifiable,
	}
}

func TestData(paths ...string) tdOpt {
	return testData(false, paths...)
}

func ModifiableTestData(paths ...string) tdOpt {
	return testData(true, paths...)
}

func projectTestData(modifiable bool, source string, dest ...string) Option {
	pathToRoot, err := testutils.GetRelativePathToProjectRoot()
	if err != nil {
		panic(err)
	}
	pathToTestdata := filepath.Join(pathToRoot, source)

	return testData(modifiable, pathToTestdata, general.OptionalDefaulted("/testdata", dest...))
}

func ProjectTestData(source string, dest ...string) Option {
	return projectTestData(false, source, dest...)
}

func ModifiableProjectTestData(source string, dest ...string) Option {
	return projectTestData(true, source, dest...)
}

func projectTestDataForCaller(modifiable bool, source string, dest ...string) Option {
	packagePath, err := pkgutils.GetPackageName(2)
	if err != nil {
		panic(err)
	}

	moduleName, err := testutils.GetModuleName()
	if err != nil {
		panic(err)
	}
	path, ok := strings.CutPrefix(packagePath, moduleName+"/")
	if !ok {
		panic("unable to find package name")
	}

	return projectTestData(modifiable, filepath.Join(path, source), dest...)
}

func ProjectTestDataForCaller(source string, dest ...string) Option {
	return projectTestDataForCaller(false, source, dest...)
}

func ModifiableProjectTestDataForCaller(source string, dest ...string) Option {
	return projectTestDataForCaller(true, source, dest...)
}

func (o tdOpt) OptionHandler() OptionHandler {
	return o
}

func (o tdOpt) Mount(cfs *composefs.ComposedFileSystem) error {
	fs, err := projectionfs.New(osfs.New(), o.source)
	if err != nil {
		return fmt.Errorf("failed to create new project fs: %w", err)
	}

	if o.modifiable {
		fs = layerfs.New(memoryfs.New(), fs)
	} else {
		fs = readonlyfs.New(fs)
	}

	if err = cfs.MkdirAll(o.path, vfs.ModePerm); err != nil {
		return err
	}

	if err := cfs.Mount(o.path, fs); err != nil {
		return fmt.Errorf("failed to mount cfs: %w", err)
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

type envOpt struct {
	DefaultOptionHandler
	env *Environment
}

func (o envOpt) OptionHandler() OptionHandler {
	return o
}

func (o envOpt) OCMContext() ocm.Context {
	return o.env.OCMContext()
}

func (o envOpt) GetFilesystem() vfs.FileSystem {
	return o.env.GetFilesystem()
}

func (o envOpt) GetFailHandler() FailHandler {
	return o.env.GetFailHandler()
}

func (o envOpt) GetEnvironment() *Environment {
	return o.env
}

/////////////////////////

type Environment struct {
	vfs.VFS
	ctx         ocm.Context
	filesystem  vfs.FileSystem
	failhandler FailHandler
}

var (
	_ Option          = (*Environment)(nil)
	_ accessio.Option = (*Environment)(nil)
)

func NewEnvironment(opts ...Option) *Environment {
	var basefs vfs.FileSystem
	var basefh FailHandler
	var ctx ocm.Context

	for _, o := range opts {
		if o == nil {
			continue
		}
		h := o.OptionHandler()
		if h == nil {
			continue
		}
		fs := h.GetFilesystem()
		if fs != nil {
			basefs = fs
		}
		fh := h.GetFailHandler()
		if fh != nil {
			basefh = fh
		}
		oc := h.OCMContext()
		if oc != nil {
			ctx = oc
		}
	}

	if basefs == nil {
		tmpfs, err := osfs.NewTempFileSystem()
		if err != nil {
			panic(err)
		}
		basefs = tmpfs
		defer func() {
			vfs.Cleanup(basefs)
		}()
	}
	if ok, err := vfs.DirExists(basefs, "/tmp"); err != nil || !ok {
		err := basefs.Mkdir("/tmp", vfs.ModePerm)
		if err != nil {
			panic(err)
		}
	}

	fs := composefs.New(basefs, "/tmp")
	for _, o := range opts {
		if o == nil {
			continue
		}
		h := o.OptionHandler()
		if h == nil {
			continue
		}
		err := h.Mount(fs)
		if err != nil {
			panic(err)
		}
	}

	if ctx == nil {
		ctx = ocm.WithCredentials(credentials.WithConfigs(config.New()).New()).New()
	}

	// TODO: delegate this to special option given for all test use cases
	vfsattr.Set(ctx.AttributesContext(), fs)
	basefs = nil

	e := &Environment{
		VFS:         vfs.New(fs),
		ctx:         ctx,
		filesystem:  fs,
		failhandler: basefh,
	}
	for _, o := range opts {
		if o == nil {
			continue
		}
		h := o.OptionHandler()
		if h == nil {
			continue
		}
		h.Propagate(e)
	}
	return e
}

func (e *Environment) OptionHandler() OptionHandler {
	return envOpt{env: e}
}

func (e *Environment) GetFilesystem() vfs.FileSystem {
	return e.FileSystem()
}

func (e *Environment) GetFailHandler() FailHandler {
	return e.failhandler
}

func (e *Environment) GetEnvironment() *Environment {
	return e
}

func (e *Environment) ApplyOption(options accessio.Options) error {
	options.SetPathFileSystem(e.FileSystem())
	return nil
}

func (e *Environment) OCMContext() ocm.Context {
	return e.ctx
}

func (e *Environment) OCIContext() oci.Context {
	return e.ctx.OCIContext()
}

func (e *Environment) CredentialsContext() credentials.Context {
	return e.ctx.CredentialsContext()
}

func (e *Environment) ConfigContext() config.Context {
	return e.ctx.ConfigContext()
}

func (e *Environment) FileSystem() vfs.FileSystem {
	return e.filesystem
}

func ExceptionFailHandler(msg string, callerSkip ...int) {
	skip := utils.Optional(callerSkip...) + 1
	st, _ := gostackparse.Parse(bytes.NewReader(debug.Stack()))
	if st == nil {
		exception.Throw(fmt.Errorf("%s", msg))
	}
	f := strings.Split(st[0].Stack[skip].Func, "/")

	exception.Throw(fmt.Errorf("%s(%d): %s", f[len(f)-1], st[0].Stack[skip+1].Line, msg))
}

// SetFailHandler sets an explicit fail handler or
// by default a fail handler throwing an exception
// is set.
func (e *Environment) SetFailHandler(h ...FailHandler) *Environment {
	e.failhandler = utils.OptionalDefaulted(FailHandler(ExceptionFailHandler), h...)
	return e
}

func (e *Environment) Fail(msg string, callerSkip ...int) {
	e.fail(msg, callerSkip...)
}

func (e *Environment) FailOnErr(err error, msg string, callerSkip ...int) {
	if msg != "" && err != nil {
		err = fmt.Errorf("%s: %w", msg, err)
	}
	e.failOn(err, callerSkip...)
}

func (e *Environment) fail(msg string, callerSkip ...int) {
	fh := e.failhandler
	if fh == nil {
		ExceptionFailHandler(msg, utils.Optional(callerSkip...)+2)
	} else {
		fh(msg, utils.Optional(callerSkip...)+2)
	}
}

func (e *Environment) failOn(err error, callerSkip ...int) {
	if err != nil {
		e.fail(err.Error(), utils.Optional(callerSkip...)+1)
	}
}
