package helmblob

import (
	"github.com/mandelsoft/goutils/optionutils"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/elements/artifactblob/api"
	base "ocm.software/ocm/api/utils/blobaccess/helm"
	"ocm.software/ocm/api/utils/misc"
)

type Option = optionutils.Option[*Options]

type Options struct {
	api.Options
	Blob base.Options
}

var (
	_ api.GeneralOptionsProvider = (*Options)(nil)
	_ Option                     = (*Options)(nil)
)

func (o *Options) ApplyTo(opts *Options) {
	o.Options.ApplyTo(&opts.Options)
	o.Blob.ApplyTo(&opts.Blob)
}

func (o *Options) Apply(opts ...Option) {
	optionutils.ApplyOptions(o, opts...)
}

////////////////////////////////////////////////////////////////////////////////
// General Options

func WithHint(h string) Option {
	return api.WrapHint[Options](h)
}

func WithGlobalAccess(a cpi.AccessSpec) Option {
	return api.WrapGlobalAccess[Options](a)
}

////////////////////////////////////////////////////////////////////////////////
// DirTree BlobAccess Options

func mapBaseOption(opts *Options) *base.Options {
	return &opts.Blob
}

func wrapBase(o base.Option) Option {
	return optionutils.OptionWrapperFunc[*base.Options, *Options](o, mapBaseOption)
}

func WithFileSystem(fs vfs.FileSystem) Option {
	return wrapBase(base.WithFileSystem(fs))
}

func WithContext(ctx oci.ContextProvider) Option {
	return wrapBase(base.WithContext(ctx))
}

func WithIVersion(v string) Option {
	return wrapBase(base.WithVersion(v))
}

func WithIVersionOverride(v string, flag ...bool) Option {
	return wrapBase(base.WithVersionOverride(v, flag...))
}

func WithCACert(v string) Option {
	return wrapBase(base.WithCACert(v))
}

func WithCACertFile(v string) Option {
	return wrapBase(base.WithCACertFile(v))
}

func WithHelmRepository(v string) Option {
	return wrapBase(base.WithHelmRepository(v))
}

func WithPrinter(v misc.Printer) Option {
	return wrapBase(base.WithPrinter(v))
}
