package wgetaccess

import "github.com/open-component-model/ocm/pkg/optionutils"

type Option = optionutils.Option[*Options]

type Options struct {
	MimeType string
}

var _ Option = (*Options)(nil)

func (o *Options) ApplyTo(opts *Options) {
	if o.MimeType != "" {
		opts.MimeType = o.MimeType
	}
}

func (o *Options) Apply(opts ...Option) {
	optionutils.ApplyOptions(o, opts...)
}

type mimetype string

func (o mimetype) ApplyTo(opts *Options) {
	opts.MimeType = string(o)
}

func WithMimeType(o string) Option {
	return mimetype((o))
}
