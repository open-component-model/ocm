// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cpi

import (
	"strings"

	"github.com/open-component-model/ocm/pkg/contexts/config/internal"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type ConfigTypeVersionScheme = runtime.TypeVersionScheme[Config, ConfigType]

func NewConfigTypeVersionScheme(kind string) ConfigTypeVersionScheme {
	return runtime.NewTypeVersionScheme[Config, ConfigType](kind, internal.NewStrictConfigTypeScheme())
}

func RegisterConfigType(rtype ConfigType) {
	internal.DefaultConfigTypeScheme.Register(rtype)
}

func RegisterConfigTypeVersions(s ConfigTypeVersionScheme) {
	internal.DefaultConfigTypeScheme.AddKnownTypes(s)
}

////////////////////////////////////////////////////////////////////////////////

type configType struct {
	runtime.VersionedTypedObjectType[Config]
	usage string
}

func NewConfigType[I Config](name string, proto I, usages ...string) ConfigType {
	return &configType{
		VersionedTypedObjectType: runtime.NewVersionedTypedObjectTypeByProto[Config, I](name, proto),
		usage:                    strings.Join(usages, "\n"),
	}
}

func NewRepositoryTypeByProtoConverter[I Config](name string, proto runtime.TypedObject, converter runtime.Converter[I, runtime.TypedObject], usages ...string) ConfigType {
	return &configType{
		VersionedTypedObjectType: runtime.NewVersionedTypedObjectTypeByProtoConverter[Config, I](name, proto, converter),
		usage:                    strings.Join(usages, "\n"),
	}
}

func NewRepositoryTypeByConverter[I Config, V runtime.TypedObject](name string, converter runtime.Converter[I, V], usages ...string) ConfigType {
	return &configType{
		VersionedTypedObjectType: runtime.NewVersionedTypedObjectTypeByConverter[Config, I](name, converter),
		usage:                    strings.Join(usages, "\n"),
	}
}

func (t *configType) Usage() string {
	return t.usage
}
