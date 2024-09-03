package ociblobaccess

import (
	"github.com/mandelsoft/goutils/optionutils"
)

type Option = optionutils.Option[*Options]

type Options struct {
	MediaType string
}

var _ Option = (*Options)(nil)

func (o *Options) ApplyTo(opts *Options) {
	if o.MediaType != "" {
		opts.MediaType = o.MediaType
	}
}

func (o *Options) Apply(opts ...Option) {
	optionutils.ApplyOptions(o, opts...)
}

////////////////////////////////////////////////////////////////////////////////
// Local options

type mediatype string

func (h mediatype) ApplyTo(opts *Options) {
	opts.MediaType = string(h)
}

func WithMediaType(h string) Option {
	return mediatype((h))
}
