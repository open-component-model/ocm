package dockerdaemonblob

import (
	"github.com/mandelsoft/goutils/optionutils"

	base "github.com/open-component-model/ocm/pkg/blobaccess/dockerdaemon"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/elements/artifactblob/api"
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
// Docker BlobAccess Options

func mapBaseOption(opts *Options) *base.Options {
	return &opts.Blob
}

func wrapBase(o base.Option) Option {
	return optionutils.OptionWrapperFunc[*base.Options, *Options](o, mapBaseOption)
}

func WithName(n string) Option {
	return wrapBase(base.WithName(n))
}

func WithVersion(v string) Option {
	return wrapBase(base.WithVersion(v))
}

func WithVersionOverride(v string, flag ...bool) Option {
	return wrapBase(base.WithVersionOverride(v, flag...))
}

func WithOrigin(o common.NameVersion) Option {
	return wrapBase(base.WithOrigin(o))
}
