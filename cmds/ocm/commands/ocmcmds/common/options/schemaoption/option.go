// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package schemaoption

import (
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/listformat"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/errors"
	utils2 "github.com/open-component-model/ocm/pkg/utils"
)

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

func New(def string, internal ...bool) *Option {
	return &Option{Defaulted: def, internal: utils2.Optional(internal...)}
}

type Option struct {
	Defaulted string
	Schema    string
	internal  bool
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Schema, "scheme", "S", o.Defaulted, "schema version")
}

func (o *Option) Complete() error {
	if o.Schema == "" {
		o.Schema = o.Defaulted
	}
	if o.Schema != "" {
		if o.Schema != compdesc.InternalSchemaVersion || !o.internal {
			s := compdesc.DefaultSchemes[o.Schema]
			if s == nil {
				s = compdesc.DefaultSchemes[metav1.GROUP+"/"+o.Schema]
				if s != nil {
					o.Schema = metav1.GROUP + "/" + o.Schema
				}
			}
			if s == nil {
				return errors.ErrUnknown(errors.KIND_SCHEMAVERSION, o.Schema)
			}
		}
	}
	return nil
}

func (o *Option) Usage() string {
	s := ""
	if o.Defaulted != "" {
		s = `
If the option <code>--scheme</code> is given, the specified component descriptor format is used/generated.
`
	} else {
		s = `
If the option <code>--scheme</code> is given, the component descriptor
is converted to the specified format for output. If no format is given
the storage format of the actual descriptor is used or, for new ones v2
is used.`
	}
	if o.internal {
		s += `
With <code>internal</code> the internal representation is shown.`
	}
	s += `
The following schema versions are supported for explicit conversions:
` + listformat.FormatList(o.Defaulted, compdesc.DefaultSchemes.Names()...)
	return s
}
