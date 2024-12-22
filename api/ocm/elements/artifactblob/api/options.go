package api

import (
	"github.com/mandelsoft/goutils/optionutils"

	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/refhints"
)

type (
	Option                 = optionutils.Option[*Options]
	GeneralOptionsProvider = optionutils.NestedOptionsProvider[*Options]
)

type Options struct {
	Global cpi.AccessSpec
	Hint   refhints.ReferenceHints
}

var (
	_ optionutils.NestedOptionsProvider[*Options] = (*Options)(nil)
	_ optionutils.Option[*Options]                = (*Options)(nil)
)

func (w *Options) NestedOptions() *Options {
	return w
}

func (o *Options) ApplyTo(opts *Options) {
	if o.Global != nil {
		opts.Global = o.Global
	}
	if o.Hint != nil {
		opts.Hint = o.Hint
	}
}

func (o *Options) Apply(opts ...Option) {
	optionutils.ApplyOptions(o, opts...)
}

type hint struct {
	hint refhints.ReferenceHints
}

func (o hint) ApplyTo(opts *Options) {
	opts.Hint = o.hint
}

func WithHint(h string) Option {
	if h == "" {
		return hint{nil}
	}
	return hint{refhints.ParseHints(h)}
}

func WithReferenceHint(h ...refhints.ReferenceHint) Option {
	return hint{h}
}

func WrapHint[O any, P optionutils.OptionTargetProvider[*Options, O]](h string) optionutils.Option[P] {
	return optionutils.OptionWrapper[*Options, O, P](WithHint(h))
}

////////////////////////////////////////////////////////////////////////////////

type global struct {
	cpi.AccessSpec
}

func (o global) ApplyTo(opts *Options) {
	opts.Global = o.AccessSpec
}

func WithGlobalAccess(a cpi.AccessSpec) Option {
	return global{a}
}

func WrapGlobalAccess[O any, P optionutils.OptionTargetProvider[*Options, O]](a cpi.AccessSpec) optionutils.Option[P] {
	return optionutils.OptionWrapper[*Options, O, P](WithGlobalAccess(a))
}
