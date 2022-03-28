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

package closureoption

import (
	"github.com/gardener/ocm/cmds/ocm/pkg/output"
	"github.com/gardener/ocm/cmds/ocm/pkg/processing"
	"github.com/spf13/pflag"
)

func From(o *output.Options) *Option {
	return o.OtherOptions.(*Option)
}

type Option struct {
	Closure bool
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&o.Closure, "closure", "c", false, "follow component references")
}

func (o *Option) Usage() string {
	return `
With the option <code>--closure</code> the complete reference tree by a component verserion is traversed.
`
}

func (o *Option) Columns(cols []string) []string {
	if o.Closure {
		return append(append(cols[:0:len(cols)+1], "REFERENCEPATH"), cols...)
	}
	return cols
}

func (o *Option) Row(path string, cols []string) []string {
	if o.Closure {
		return append(append(cols[:0:len(cols)+1], path), cols...)
	}
	return cols
}

func (o *Option) RowEnricher(path func(interface{}) string, row func(interface{}) []string) func(interface{}) []string {
	if o.Closure {
		return func(e interface{}) []string {
			cols := row(e)
			return append(append(cols[:0:len(cols)+1], path(e)), cols...)
		}
	}
	return row
}

func (o *Option) Mapper(path func(interface{}) string, mapper processing.MappingFunction) processing.MappingFunction {
	if o.Closure {
		return func(e interface{}) interface{} {
			cols := mapper(e).([]string)
			return append(append(cols[:0:len(cols)+1], path(e)), cols...)
		}
	}
	return mapper
}
