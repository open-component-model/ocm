package helm

import (
	"github.com/mandelsoft/goutils/optionutils"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/utils"
	common "ocm.software/ocm/api/utils/misc"
)

type Option = optionutils.Option[*Options]

type Options struct {
	Context         oci.Context
	FileSystem      vfs.FileSystem
	Version         string
	OverrideVersion *bool
	HelmRepository  string
	CACert          string
	CACertFile      string

	Printer common.Printer
}

func (o *Options) ApplyTo(opts *Options) {
	if opts == nil {
		return
	}
	if o.Context != nil {
		opts.Context = o.Context
	}
	if o.FileSystem != nil {
		opts.FileSystem = o.FileSystem
	}
	if o.Version != "" {
		opts.Version = o.Version
	}
	if o.OverrideVersion != nil {
		opts.OverrideVersion = o.OverrideVersion
	}
	if o.HelmRepository != "" {
		opts.HelmRepository = o.HelmRepository
	}
	if o.CACert != "" {
		opts.CACert = o.CACert
	}
	if o.CACertFile != "" {
		opts.CACertFile = o.CACertFile
	}
	if o.Printer != nil {
		opts.Printer = o.Printer
	}
}

func (o *Options) OCIContext() oci.Context {
	if o.Context == nil {
		return oci.DefaultContext()
	}
	return o.Context
}

////////////////////////////////////////////////////////////////////////////////

type context struct {
	oci.Context
}

func (o context) ApplyTo(opts *Options) {
	opts.Context = o
}

func WithContext(ctx oci.ContextProvider) Option {
	return context{ctx.OCIContext()}
}

////////////////////////////////////////////////////////////////////////////////

type fileSystem struct {
	fs vfs.FileSystem
}

func (o *fileSystem) ApplyTo(opts *Options) {
	opts.FileSystem = o.fs
}

func WithFileSystem(fs vfs.FileSystem) Option {
	return &fileSystem{fs: fs}
}

////////////////////////////////////////////////////////////////////////////////

type version string

func (o version) ApplyTo(opts *Options) {
	opts.Version = string(o)
}

func WithVersion(v string) Option {
	return version(v)
}

////////////////////////////////////////////////////////////////////////////////

type override struct {
	flag    bool
	version string
}

func (o *override) ApplyTo(opts *Options) {
	opts.OverrideVersion = utils.BoolP(o.flag)
	opts.Version = o.version
}

func WithVersionOverride(v string, flag ...bool) Option {
	return &override{
		version: v,
		flag:    utils.OptionalDefaultedBool(true, flag...),
	}
}

////////////////////////////////////////////////////////////////////////////////

type helmrepo string

func (o helmrepo) ApplyTo(opts *Options) {
	opts.HelmRepository = string(o)
}

// WithHelmRepository defines the helm repository to read from.
func WithHelmRepository(v string) Option {
	return helmrepo(v)
}

////////////////////////////////////////////////////////////////////////////////

type cacert string

func (o cacert) ApplyTo(opts *Options) {
	opts.CACert = string(o)
}

func WithCACert(v string) Option {
	return cacert(v)
}

////////////////////////////////////////////////////////////////////////////////

type cacertfile string

func (o cacertfile) ApplyTo(opts *Options) {
	opts.CACertFile = string(o)
}

func WithCACertFile(v string) Option {
	return cacertfile(v)
}

////////////////////////////////////////////////////////////////////////////////

type printer struct {
	common.Printer
}

func (o printer) ApplyTo(opts *Options) {
	opts.Printer = o
}

func WithPrinter(p common.Printer) Option {
	return printer{p}
}
