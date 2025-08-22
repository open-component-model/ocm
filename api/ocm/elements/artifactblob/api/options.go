package api

import (
	"ocm.software/ocm/api/ocm/cpi"
)

type Option interface {
	ApplyTo(opts *Options)
}

type OptionFunc func(opts *Options)

func (f OptionFunc) ApplyTo(opts *Options) {
	f(opts)
}

type GeneralOptionsProvider interface {
	NestedOptions() *Options
}

type Options struct {
	Global cpi.AccessSpec
	Hint   string
}

var (
	_ GeneralOptionsProvider = (*Options)(nil)
	_ Option                 = (*Options)(nil)
)

func (w *Options) NestedOptions() *Options {
	return w
}

func (o *Options) ApplyTo(opts *Options) {
	if opts == nil {
		return
	}
	if o.Global != nil {
		opts.Global = o.Global
	}
	if o.Hint != "" {
		opts.Hint = o.Hint
	}
}

func (o *Options) Apply(opts ...Option) {
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyTo(o)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// Local Options

func WithHint(h string) Option {
	return OptionFunc(func(opts *Options) {
		opts.Hint = h
	})
}

func WithGlobalAccess(a cpi.AccessSpec) Option {
	return OptionFunc(func(opts *Options) {
		opts.Global = a
	})
}
