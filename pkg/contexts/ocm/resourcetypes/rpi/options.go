// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package rpi

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
)

type GeneralOptionsProvider interface {
	GeneralOptions() *Options
}

type Option = ResourceOption[*Options]

type Options struct {
	Global cpi.AccessSpec
	Hint   string
}

func (w *Options) GeneralOptions() *Options {
	return w
}

func (o *Options) ApplyTo(opts *Options) {
	if o.Global != nil {
		opts.Global = o.Global
	}
	if o.Hint != "" {
		opts.Hint = o.Hint
	}
}

type hint string

func (o hint) ApplyTo(opts *Options) {
	opts.Hint = string(o)
}

func WithHint(h string) Option {
	return hint(h)
}

func WrapHint[O any, P OptionTargetProvider[O]](h string) ResourceOption[P] {
	return OptionWrapper[O, P](WithHint(h))
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

func WrapGlobalAccess[O any, P OptionTargetProvider[O]](a cpi.AccessSpec) ResourceOption[P] {
	return OptionWrapper[O, P](WithGlobalAccess(a))
}

////////////////////////////////////////////////////////////////////////////////

type ResourceOption[T any] interface {
	ApplyTo(T)
}

type OptionTargetProvider[O any] interface {
	GeneralOptionsProvider
	*O
}

func OptionWrapper[O any, P OptionTargetProvider[O]](o Option) ResourceOption[P] {
	return optionWrapper[O, P]{o}
}

type optionWrapper[O any, P OptionTargetProvider[O]] struct {
	opt Option
}

func (w optionWrapper[O, P]) ApplyTo(opts P) {
	w.opt.ApplyTo(opts.GeneralOptions())
}

////////////////////////////////////////////////////////////////////////////////

func EvalOptions[O any](opts ...ResourceOption[*O]) *O {
	var eff O
	for _, opt := range opts {
		opt.ApplyTo(&eff)
	}
	return &eff
}
