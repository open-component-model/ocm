// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package options

import (
	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
)

const (
	TYPE_STRING        = "string"
	TYPE_STRINGARRAY   = "[]string"
	TYPE_STRING2STRING = "string=string"
	TYPE_INT           = "int"
	TYPE_BOOL          = "bool"
	TYPE_YAML          = "YAML"
	TYPE_STRINGMAP     = "map[string]YAML"
	TYPE_STRING2YAML   = "string=YAML"
)

func init() {
	DefaultRegistry.RegisterType(TYPE_STRING, flagsets.NewStringOptionType, "string value")
	DefaultRegistry.RegisterType(TYPE_STRINGARRAY, flagsets.NewStringArrayOptionType, "list of string values")
	DefaultRegistry.RegisterType(TYPE_STRING2STRING, flagsets.NewStringMapOptionType, "string map defined by dedicated assignments")
	DefaultRegistry.RegisterType(TYPE_INT, flagsets.NewIntOptionType, "integer value")
	DefaultRegistry.RegisterType(TYPE_BOOL, flagsets.NewBoolOptionType, "boolean flag")
	DefaultRegistry.RegisterType(TYPE_YAML, flagsets.NewYAMLOptionType, "JSON or YAML document string")
	DefaultRegistry.RegisterType(TYPE_STRINGMAP, flagsets.NewValueMapYAMLOptionType, "JSON or YAML map")
	DefaultRegistry.RegisterType(TYPE_STRING2YAML, flagsets.NewValueMapOptionType, "string map with arbitrary values defined by dedicated assignments")
}

func RegisterOption(o flagsets.ConfigOptionType) flagsets.ConfigOptionType {
	DefaultRegistry.RegisterOption(o)
	return o
}
