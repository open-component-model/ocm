// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package failonerroroption

import (
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
)

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

func New() *Option {
	return &Option{}
}

type Option struct {
	Fail bool
	err  error
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&o.Fail, "fail-on-error", "", false, "fail on validation error")
}

var _ options.Options = (*Option)(nil)

func (o *Option) GetError() error {
	return o.err
}
func (o *Option) SetError(err error) {
	o.err = err
}

func (o *Option) AddError(err error) {
	if err == nil {
		return
	}
	if o.err == nil {
		o.err = errors.ErrList().Add(err)
	} else {
		if l, ok := o.err.(*errors.ErrorList); ok {
			l.Add(err)
		} else {
			o.err = errors.ErrList().Add(o.err, err)
		}
	}
}

func (o *Option) ActivatedError() error {
	if o.Fail {
		return o.err
	}
	return nil
}
