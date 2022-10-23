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

package flagsets

import (
	"reflect"

	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/pkg/cobrautils/flags"
)

type TypeOptionBase struct {
	name        string
	description string
}

func (b *TypeOptionBase) Name() string {
	return b.name
}

func (b *TypeOptionBase) Description() string {
	return b.description
}

////////////////////////////////////////////////////////////////////////////////

type OptionBase struct {
	otyp ConfigOptionType
}

func NewOptionBase(otyp ConfigOptionType) OptionBase {
	return OptionBase{otyp: otyp}
}

func (b OptionBase) Type() ConfigOptionType {
	return b.otyp
}

func (b OptionBase) Name() string {
	return b.otyp.Name()
}

func (b OptionBase) Description() string {
	return b.otyp.Description()
}

////////////////////////////////////////////////////////////////////////////////

type StringOptionType struct {
	TypeOptionBase
}

func NewStringOptionType(name string, description string) ConfigOptionType {
	return &StringOptionType{
		TypeOptionBase: TypeOptionBase{name, description},
	}
}

type StringOption struct {
	OptionBase
	value string
}

var _ Option = (*StringOption)(nil)

func (o *StringOption) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.value, o.otyp.Name(), "", "", o.otyp.Description())
}

func (o *StringOption) Value() interface{} {
	return o.value
}

func (s *StringOptionType) Equal(optionType ConfigOptionType) bool {
	return reflect.DeepEqual(s, optionType)
}

func (s *StringOptionType) Create() Option {
	return &StringOption{
		OptionBase: NewOptionBase(s),
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type StringArrayOptionType struct {
	TypeOptionBase
}

func NewStringArrayOptionType(name string, description string) ConfigOptionType {
	return &StringArrayOptionType{
		TypeOptionBase: TypeOptionBase{name, description},
	}
}

type StringArrayOption struct {
	OptionBase
	value []string
}

var _ Option = (*StringArrayOption)(nil)

func (o *StringArrayOption) AddFlags(fs *pflag.FlagSet) {
	fs.StringArrayVarP(&o.value, o.otyp.Name(), "", nil, o.otyp.Description())
}

func (o *StringArrayOption) Value() interface{} {
	return o.value
}

func (s *StringArrayOptionType) Equal(optionType ConfigOptionType) bool {
	return reflect.DeepEqual(s, optionType)
}

func (s *StringArrayOptionType) Create() Option {
	return &StringArrayOption{
		OptionBase: NewOptionBase(s),
	}
}

////////////////////////////////////////////////////////////////////////////////

type BoolOptionType struct {
	TypeOptionBase
}

func NewBoolOptionType(name string, description string) ConfigOptionType {
	return &BoolOptionType{
		TypeOptionBase: TypeOptionBase{name, description},
	}
}

type BoolOption struct {
	OptionBase
	value bool
}

var _ Option = (*BoolOption)(nil)

func (o *BoolOption) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&o.value, o.otyp.Name(), "", false, o.otyp.Description())
}

func (o *BoolOption) Value() interface{} {
	return o.value
}

func (s *BoolOptionType) Equal(optionType ConfigOptionType) bool {
	return reflect.DeepEqual(s, optionType)
}

func (s *BoolOptionType) Create() Option {
	return &BoolOption{
		OptionBase: NewOptionBase(s),
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type IntOptionType struct {
	TypeOptionBase
}

func NewIntOptionType(name string, description string) ConfigOptionType {
	return &IntOptionType{
		TypeOptionBase: TypeOptionBase{name, description},
	}
}

type IntOption struct {
	OptionBase
	value int
}

var _ Option = (*IntOption)(nil)

func (o *IntOption) AddFlags(fs *pflag.FlagSet) {
	fs.IntVarP(&o.value, o.otyp.Name(), "", 0, o.otyp.Description())
}

func (o *IntOption) Value() interface{} {
	return o.value
}

func (s *IntOptionType) Equal(optionType ConfigOptionType) bool {
	return reflect.DeepEqual(s, optionType)
}

func (s *IntOptionType) Create() Option {
	return &IntOption{
		OptionBase: NewOptionBase(s),
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type YAMLOptionType struct {
	TypeOptionBase
}

func NewYAMLOptionType(name string, description string) ConfigOptionType {
	return &YAMLOptionType{
		TypeOptionBase: TypeOptionBase{name, description},
	}
}

type YAMLOption struct {
	OptionBase
	value interface{}
}

var _ Option = (*YAMLOption)(nil)

func (o *YAMLOption) AddFlags(fs *pflag.FlagSet) {
	flags.YAMLVarP(fs, &o.value, o.otyp.Name(), "", nil, o.otyp.Description())
}

func (o *YAMLOption) Value() interface{} {
	return o.value
}

func (s *YAMLOptionType) Equal(optionType ConfigOptionType) bool {
	return reflect.DeepEqual(s, optionType)
}

func (s *YAMLOptionType) Create() Option {
	return &YAMLOption{
		OptionBase: NewOptionBase(s),
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type ValueMapOptionType struct {
	TypeOptionBase
}

func NewValueMapOptionType(name string, description string) ConfigOptionType {
	return &ValueMapOptionType{
		TypeOptionBase: TypeOptionBase{name, description},
	}
}

type ValueMapOption struct {
	OptionBase
	value map[string]interface{}
}

var _ Option = (*ValueMapOption)(nil)

func (o *ValueMapOption) AddFlags(fs *pflag.FlagSet) {
	flags.YAMLVarP(fs, &o.value, o.otyp.Name(), "", nil, o.otyp.Description())
}

func (o *ValueMapOption) Value() interface{} {
	return o.value
}

func (s *ValueMapOptionType) Equal(optionType ConfigOptionType) bool {
	return reflect.DeepEqual(s, optionType)
}

func (s *ValueMapOptionType) Create() Option {
	return &YAMLOption{
		OptionBase: NewOptionBase(s),
	}
}
