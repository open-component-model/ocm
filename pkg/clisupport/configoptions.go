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

package clisupport

import (
	"fmt"

	"github.com/spf13/pflag"
)

type Option interface {
	Name() string
	AddFlags(fs *pflag.FlagSet)
	Value() interface{}
}

type ConfigOptions interface {
	AddFlags(fs *pflag.FlagSet)
	Check(set ConfigOptionTypeSet, desc string) error
	GetValue(name string) (interface{}, bool)
	Changed() bool

	FilterBy(func(name string) bool) ConfigOptions
}

type configOptions struct {
	options []Option
	flags   *pflag.FlagSet
}

func NewOptions(opts []Option) ConfigOptions {
	return &configOptions{options: opts}
}

func (o *configOptions) GetValue(name string) (interface{}, bool) {
	for _, opt := range o.options {
		if opt.Name() == name {
			return opt.Value(), o.flags.Changed(name)
		}
	}
	return nil, false
}

func (o *configOptions) AddFlags(fs *pflag.FlagSet) {
	for _, opt := range o.options {
		opt.AddFlags(fs)
	}
	o.flags = fs
}

func (o *configOptions) Changed() bool {
	for _, opt := range o.options {
		if o.flags.Changed(opt.Name()) {
			return true
		}
	}
	return false
}

func (o *configOptions) FilterBy(filter func(name string) bool) ConfigOptions {
	if filter == nil {
		return o
	}
	var options []Option

	for _, opt := range o.options {
		if filter(opt.Name()) {
			options = append(options, opt)
		}
	}
	return &configOptions{
		options: options,
		flags:   o.flags,
	}
}

func (o *configOptions) Check(set ConfigOptionTypeSet, desc string) error {
	if desc != "" {
		desc = " for " + desc
	}

	if set == nil {
		for _, opt := range o.options {
			if o.flags.Changed(opt.Name()) {
				return fmt.Errorf("option %q given, but not possible%s", opt.Name(), desc)
			}
		}
	} else {
		for _, opt := range o.options {
			if o.flags.Changed(opt.Name()) && set.GetOptionType(opt.Name()) == nil {
				if desc == "" {
					return fmt.Errorf("option %q given, but not valid for option set %q", opt.Name(), set.Name())
				}
				return fmt.Errorf("option %q given, but not possible%s", opt.Name(), desc)
			}
		}
	}
	return nil
}
