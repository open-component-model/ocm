// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package inputs

import (
	"fmt"
	"strings"

	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
	"github.com/open-component-model/ocm/pkg/utils"
)

type FieldSetter func(opts flagsets.ConfigOptions, opt flagsets.ConfigOptionType, config flagsets.Config) error

type InputField interface {
	OptionType() flagsets.ConfigOptionType
	Usage() string
	AddConfig(opts flagsets.ConfigOptions, config flagsets.Config) error
}

type inputField struct {
	name   string
	desc   string
	opt    flagsets.ConfigOptionType
	setter FieldSetter
}

func NewInputField(name, desc string, opt flagsets.ConfigOptionType, setter FieldSetter) InputField {
	return &inputField{
		name:   name,
		desc:   desc,
		opt:    opt,
		setter: setter,
	}
}

func (f *inputField) OptionType() flagsets.ConfigOptionType {
	return f.opt
}

func (f *inputField) AddConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	return f.setter(opts, f.opt, config)
}

func (f *inputField) Usage() string {
	return InputFieldDoc(f.name, f.desc, f.opt)
}

func (f *inputField) InputFields() []InputField {
	return []InputField{f}
}

type InputFields []InputField

type InputFieldSource interface {
	InputFields() []InputField
}

func NewInputFields(fields ...InputFieldSource) InputFields {
	var list []InputField
	for _, f := range fields {
		list = append(list, f.InputFields()...)
	}
	return list
}

func (i InputFields) InputFields() []InputField {
	return i
}

func (i InputFields) Usage() string {
	usage := ""
	for _, f := range i {
		usage += f.Usage()
	}
	return usage
}

func (i InputFields) AddConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	for _, f := range i {
		err := f.AddConfig(opts, config)
		if err != nil {
			return err
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func InputFieldDoc(name string, desc string, opt flagsets.ConfigOptionType) string {
	oname := ""
	if opt != nil {
		oname = fmt.Sprintf(" (<code>%s</code>)", opt.GetName())
	}
	return `- **<code>` + name + `</code>** *string* ` + oname + `

` + utils.IndentLines(strings.TrimSpace(desc), "  ", false) + `
`
}
