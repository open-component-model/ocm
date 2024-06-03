package githubaccess

import (
	"github.com/mandelsoft/goutils/optionutils"
)

type Option = optionutils.Option[*Options]

type Options struct {
	APIHostName string
}

var _ Option = (*Options)(nil)

func (o *Options) ApplyTo(opts *Options) {
	if o.APIHostName != "" {
		opts.APIHostName = o.APIHostName
	}
}

func (o *Options) Apply(opts ...Option) {
	optionutils.ApplyOptions(o, opts...)
}

////////////////////////////////////////////////////////////////////////////////
// Local options

type apihostname string

func (h apihostname) ApplyTo(opts *Options) {
	opts.APIHostName = string(h)
}

func WithAPIHostName(h string) Option {
	return apihostname((h))
}
