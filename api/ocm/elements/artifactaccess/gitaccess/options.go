package gitaccess

import (
	"github.com/mandelsoft/goutils/optionutils"
)

type Option = optionutils.Option[*Options]

type Options struct {
	URL    string
	Ref    string
	Commit string
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

type commitSpec string

func (h commitSpec) ApplyTo(opts *Options) {
	opts.Commit = string(h)
}

func WithCommit(c string) Option {
	return commitSpec(c)
}
