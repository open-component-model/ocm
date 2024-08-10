package githubaccess

import (
	"github.com/mandelsoft/goutils/optionutils"
)

type Option = optionutils.Option[*Options]

type Options struct {
	URL      string
	Ref      string
	PathSpec string
}

var _ Option = (*Options)(nil)

func (o *Options) ApplyTo(opts *Options) {
	if o.URL != "" {
		opts.URL = o.URL
	}
}

func (o *Options) Apply(opts ...Option) {
	optionutils.ApplyOptions(o, opts...)
}

// //////////////////////////////////////////////////////////////////////////////
// Local options

type url string

func (h url) ApplyTo(opts *Options) {
	opts.URL = string(h)
}

func WithURL(h string) Option {
	return url(h)
}

type ref string

func (h ref) ApplyTo(opts *Options) {
	opts.Ref = string(h)
}

func WithRef(h string) Option {
	return ref(h)
}

type pathSpec string

func (h pathSpec) ApplyTo(opts *Options) {
	opts.PathSpec = string(h)
}

func WithPathSpec(h string) Option {
	return pathSpec(h)
}
