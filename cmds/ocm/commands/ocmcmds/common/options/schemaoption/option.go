// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package schemaoption

import (
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/errors"
)

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

func New(def string) *Option {
	return &Option{Defaulted: def}
}

type Option struct {
	Defaulted string
	Schema    string
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Schema, "scheme", "S", o.Defaulted, "schema version")
}

func (o *Option) Complete() error {
	if o.Schema == "" {
		o.Schema = o.Defaulted
	}
	if o.Schema != "" {
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
	return nil
}

func (o *Option) Usage() string {
	s := ""
	if o.Defaulted != "" {
		s = `
It the option <code>--scheme</code> is given, the given component descriptor format is used/generated.
`
	} else {
		s = `
It the option <code>--scheme</code> is given, the given component descriptor is converted to given format for output.
`
	}
	s += `The following schema versions are supported:
` + utils.FormatList(o.Defaulted, compdesc.DefaultSchemes.Names()...)
	return s
}
