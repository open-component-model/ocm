// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package flagsets

import (
	"reflect"

	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/pkg/cobrautils/flag"
	"github.com/open-component-model/ocm/pkg/cobrautils/groups"
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
	otyp   ConfigOptionType
	flag   *pflag.Flag
	groups []string
}

func NewOptionBase(otyp ConfigOptionType) OptionBase {
	return OptionBase{otyp: otyp}
}

func (b *OptionBase) Type() ConfigOptionType {
	return b.otyp
}

func (b *OptionBase) Name() string {
	return b.otyp.Name()
}

func (b *OptionBase) Description() string {
	return b.otyp.Description()
}

func (b *OptionBase) Changed() bool {
	return b.flag.Changed
}

func (b *OptionBase) AddGroups(groups ...string) {
	b.groups = AddGroups(b.groups, groups...)
	b.addGroups()
}

func (b *OptionBase) addGroups() {
	if len(b.groups) == 0 || b.flag == nil {
		return
	}
	if b.flag.Annotations == nil {
		b.flag.Annotations = map[string][]string{}
	}
	list := b.flag.Annotations[groups.FlagGroupAnnotation]
	b.flag.Annotations[groups.FlagGroupAnnotation] = AddGroups(list, b.groups...)
}

func (b *OptionBase) TweakFlag(f *pflag.Flag) {
	b.flag = f
	b.addGroups()
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
	o.TweakFlag(fs.StringVarPF(&o.value, o.otyp.Name(), "", "", o.otyp.Description()))
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
	o.TweakFlag(fs.StringArrayVarPF(&o.value, o.otyp.Name(), "", nil, o.otyp.Description()))
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
	o.TweakFlag(fs.BoolVarPF(&o.value, o.otyp.Name(), "", false, o.otyp.Description()))
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
	o.TweakFlag(fs.IntVarPF(&o.value, o.otyp.Name(), "", 0, o.otyp.Description()))
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
	o.TweakFlag(flag.YAMLVarPF(fs, &o.value, o.otyp.Name(), "", nil, o.otyp.Description()))
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

type ValueMapYAMLOptionType struct {
	TypeOptionBase
}

func NewValueMapYAMLOptionType(name string, description string) ConfigOptionType {
	return &ValueMapYAMLOptionType{
		TypeOptionBase: TypeOptionBase{name, description},
	}
}

type ValueMapYAMLOption struct {
	OptionBase
	value map[string]interface{}
}

var _ Option = (*ValueMapYAMLOption)(nil)

func (o *ValueMapYAMLOption) AddFlags(fs *pflag.FlagSet) {
	o.TweakFlag(flag.YAMLVarPF(fs, &o.value, o.otyp.Name(), "", nil, o.otyp.Description()))
}

func (o *ValueMapYAMLOption) Value() interface{} {
	return o.value
}

func (s *ValueMapYAMLOptionType) Equal(optionType ConfigOptionType) bool {
	return reflect.DeepEqual(s, optionType)
}

func (s *ValueMapYAMLOptionType) Create() Option {
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
	o.TweakFlag(flag.StringToValueVarPF(fs, &o.value, o.otyp.Name(), "", nil, o.otyp.Description()))
}

func (o *ValueMapOption) Value() interface{} {
	return o.value
}

func (s *ValueMapOptionType) Equal(optionType ConfigOptionType) bool {
	return reflect.DeepEqual(s, optionType)
}

func (s *ValueMapOptionType) Create() Option {
	return &ValueMapOption{
		OptionBase: NewOptionBase(s),
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type StringMapOptionType struct {
	TypeOptionBase
}

func NewStringMapOptionType(name string, description string) ConfigOptionType {
	return &StringMapOptionType{
		TypeOptionBase: TypeOptionBase{name, description},
	}
}

type StringMapOption struct {
	OptionBase
	value map[string]string
}

var _ Option = (*StringMapOption)(nil)

func (o *StringMapOption) AddFlags(fs *pflag.FlagSet) {
	o.TweakFlag(flag.StringToStringVarPF(fs, &o.value, o.otyp.Name(), "", nil, o.otyp.Description()))
}

func (o *StringMapOption) Value() interface{} {
	return o.value
}

func (s *StringMapOptionType) Equal(optionType ConfigOptionType) bool {
	return reflect.DeepEqual(s, optionType)
}

func (s *StringMapOptionType) Create() Option {
	return &StringMapOption{
		OptionBase: NewOptionBase(s),
	}
}
