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

package options

import (
	"reflect"

	"github.com/gardener/ocm/cmds/ocm/clictx"
	"github.com/gardener/ocm/cmds/ocm/pkg/output/out"
	"github.com/spf13/pflag"
)

type OptionsProcessor func(Options) error

type Complete interface {
	Complete() error
}

type CompleteWithOutputContext interface {
	Complete(ctx out.Context) error
}

type CompleteWithCLIContext interface {
	Complete(ctx clictx.Context) error
}

type Usage interface {
	Usage() string
}

type Options interface {
	AddFlags(fs *pflag.FlagSet)
}

////////////////////////////////////////////////////////////////////////////////

type OptionSet []Options

type OptionSetProvider interface {
	AsOptionSet() OptionSet
}

func (s OptionSet) AddFlags(fs *pflag.FlagSet) {
	for _, o := range s {
		o.AddFlags(fs)
	}
}

func (s OptionSet) AsOptionSet() OptionSet {
	return s
}

func (s OptionSet) Usage() string {
	u := ""
	for _, n := range s {
		if c, ok := n.(Usage); ok {
			u += c.Usage()
		}
	}
	return u
}

func (s OptionSet) Options(proto Options) interface{} {
	t := reflect.TypeOf(proto)
	for _, o := range s {
		if reflect.TypeOf(o) == t {
			return o
		}
		if set, ok := o.(OptionSetProvider); ok {
			r := set.AsOptionSet().Options(proto)
			if r != nil {
				return r
			}
		}
	}
	return nil
}

// Get extracts the option for a given target. This might be a
// - pointer to a struct implementing the Options interface which
//   will fill the struct with a copy of the options OR
// - a pointer to such a pointer which will be filled with the
//   pointer to the actual member of the OptionSet.
func (s OptionSet) Get(proto interface{}) bool {
	val := true
	t := reflect.TypeOf(proto)
	if t.Elem().Kind() == reflect.Ptr {
		t = t.Elem()
		val = false
	}
	for _, o := range s {
		if reflect.TypeOf(o) == t {
			if val {
				reflect.ValueOf(proto).Elem().Set(reflect.ValueOf(o).Elem())
			} else {
				reflect.ValueOf(proto).Elem().Set(reflect.ValueOf(o))
			}
			return true
		}
		if set, ok := o.(OptionSetProvider); ok {
			r := set.AsOptionSet().Get(proto)
			if r {
				return r
			}
		}
	}
	return false
}

func (s OptionSet) ProcessOnOptions(f OptionsProcessor) error {
	for _, n := range s {
		err := f(n)
		if err != nil {
			return err
		}
		if set, ok := n.(OptionSetProvider); ok {
			err = set.AsOptionSet().ProcessOnOptions(f)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func CompleteOptions(opt Options) error {
	if c, ok := opt.(Complete); ok {
		return c.Complete()
	}
	return nil
}

func CompleteOptionsWithCLIContext(ctx clictx.Context) OptionsProcessor {
	return func(opt Options) error {
		if c, ok := opt.(CompleteWithCLIContext); ok {
			return c.Complete(ctx)
		}
		if c, ok := opt.(CompleteWithOutputContext); ok {
			return c.Complete(ctx)
		}
		return CompleteOptions(opt)
	}
}
