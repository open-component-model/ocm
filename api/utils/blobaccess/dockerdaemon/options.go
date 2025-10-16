package dockerdaemon

import (
	"github.com/mandelsoft/goutils/optionutils"

	cpi "ocm.software/ocm/api/oci/types"
	"ocm.software/ocm/api/utils"
	common "ocm.software/ocm/api/utils/misc"
)

type Option = optionutils.Option[*Options]

type Options struct {
	Context         cpi.Context
	Name            string
	Version         string
	OverrideVersion *bool
	Origin          *common.NameVersion
}

func (o *Options) ApplyTo(opts *Options) {
	if opts == nil {
		return
	}
	if o.Context != nil {
		opts.Context = o.Context
	}
	if o.Name != "" {
		opts.Name = o.Name
	}
	if o.Version != "" {
		opts.Version = o.Version
	}
	if o.OverrideVersion != nil {
		opts.OverrideVersion = o.OverrideVersion
	}
	if o.Origin != nil {
		opts.Origin = o.Origin
	}
}

////////////////////////////////////////////////////////////////////////////////

type context struct {
	cpi.Context
}

func (o context) ApplyTo(opts *Options) {
	opts.Context = o
}

func WithContext(ctx cpi.ContextProvider) Option {
	return context{ctx.OCIContext()}
}

////////////////////////////////////////////////////////////////////////////////

type name string

func (o name) ApplyTo(opts *Options) {
	opts.Name = string(o)
}

func WithName(n string) Option {
	return name(n)
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

type compvers common.NameVersion

func (o compvers) ApplyTo(opts *Options) {
	n := common.NameVersion(o)
	opts.Origin = &n
}

func WithOrigin(o common.NameVersion) Option {
	return compvers(o)
}
