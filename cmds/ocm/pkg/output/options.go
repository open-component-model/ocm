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

package output

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gardener/ocm/cmds/ocm/clictx"
	"github.com/gardener/ocm/cmds/ocm/pkg/options"
	"github.com/gardener/ocm/cmds/ocm/pkg/output/out"
	"github.com/spf13/pflag"
)

func From(o options.OptionSetProvider) *Options {
	var opts *Options
	o.AsOptionSet().Get(&opts)
	return opts
}

type Options struct {
	options.OptionSet

	Outputs     Outputs
	output      string
	Output      *string
	Sort        []string
	FixedColums int
	Context     out.Context
}

func OutputOptions(outputs Outputs, opts ...options.Options) *Options {
	return &Options{
		Outputs:   outputs,
		OptionSet: opts,
	}
}

func (o *Options) Options(proto options.Options) interface{} {
	return o.OptionSet.Options(proto)
}

func (o *Options) Get(proto interface{}) bool {
	return o.OptionSet.Get(proto)
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	s := ""
	if len(o.Outputs) > 1 {
		list := []string{}
		for o := range o.Outputs {
			list = append(list, o)
		}
		sort.Strings(list)
		sep := ""
		for _, o := range list {
			if o != "" {
				s = fmt.Sprintf("%s%s%s", s, sep, o)
				sep = ", "
			}
		}
		fs.StringVarP(&o.output, "output", "o", "", fmt.Sprintf("output mode (%s)", s))
	}
	fs.StringArrayVarP(&o.Sort, "sort", "s", nil, "sort fields")

	o.OptionSet.AddFlags(fs)
}

func (o *Options) Complete(ctx clictx.Context) error {
	o.Context = ctx
	if o.output != "" {
		o.Output = &o.output
	}
	var fields []string

	for _, f := range o.Sort {
		split := strings.Split(f, ",")
		for _, s := range split {
			s = strings.TrimSpace(s)
			if s != "" {
				fields = append(fields, s)
			}
		}
	}
	o.Sort = fields
	err := o.ProcessOnOptions(options.CompleteOptions)
	if err != nil {
		return err
	}
	err = o.ProcessOnOptions(options.CompleteOptionsWithCLIContext(ctx))
	if err != nil {
		return err
	}
	return err
}

func (o *Options) Usage() string {
	s := o.OptionSet.Usage()

	if len(o.Outputs) > 1 {
		s += `
With the option <code>--output</code> the out put mode can be selected.
The following modes are supported:
`
		list := []string{}
		for o := range o.Outputs {
			list = append(list, o)
		}
		sort.Strings(list)
		for _, m := range list {
			if m != "" {
				s += " - " + m + "\n"
			}
		}
	}
	return s
}

func (o *Options) Create() (Output, error) {
	return o.Outputs.Create(o)
}

////////////////////////////////////////////////////////////////////////////////
