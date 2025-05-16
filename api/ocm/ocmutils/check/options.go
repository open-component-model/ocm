package check

import (
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/goutils/optionutils"
)

type Option = optionutils.Option[*Options]

type Options struct {
	CheckLocalResources *bool
	CheckLocalSources   *bool
}

var _ Option = (*Options)(nil)

func (o *Options) ApplyTo(opts *Options) {
	optionutils.ApplyOption(o.CheckLocalResources, &opts.CheckLocalResources)
	optionutils.ApplyOption(o.CheckLocalSources, &opts.CheckLocalSources)
}

////////////////////////////////////////////////////////////////////////////////

type localSources bool

func LocalSourcesOnly(b ...bool) Option {
	return localSources(general.OptionalDefaultedBool(true, b...))
}

func (l localSources) ApplyTo(t *Options) {
	t.CheckLocalSources = generics.PointerTo(bool(l))
}

////////////////////////////////////////////////////////////////////////////////

type localResources bool

func LocalResourcesOnly(b ...bool) Option {
	return localResources(general.OptionalDefaultedBool(true, b...))
}

func (l localResources) ApplyTo(t *Options) {
	t.CheckLocalResources = generics.PointerTo(bool(l))
}
