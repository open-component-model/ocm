package ociartifactblob

import (
	"github.com/mandelsoft/goutils/optionutils"

	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/elements/artifactblob/api"
	base "ocm.software/ocm/api/utils/blobaccess/ociartifact"
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

func WithContext(ctx oci.ContextProvider) Option {
	return wrapBase(base.WithContext(ctx))
}

func WithVersion(v string) Option {
	return wrapBase(base.WithVersion(v))
}

func WithPrinter(v misc.Printer) Option {
	return wrapBase(base.WithPrinter(v))
}
