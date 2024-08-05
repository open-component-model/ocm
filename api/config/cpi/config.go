package cpi

import (
	"strings"

	"ocm.software/ocm/api/config/internal"
	"ocm.software/ocm/api/utils/runtime"
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

func NewConfigType[I Config](name string, usages ...string) ConfigType {
	return &configType{
		VersionedTypedObjectType: runtime.NewVersionedTypedObjectType[Config, I](name),
		usage:                    strings.Join(usages, "\n"),
	}
}

func NewConfigTypeyConverter[I Config, V runtime.TypedObject](name string, converter runtime.Converter[I, V], usages ...string) ConfigType {
	return &configType{
		VersionedTypedObjectType: runtime.NewVersionedTypedObjectTypeByConverter[Config, I](name, converter),
		usage:                    strings.Join(usages, "\n"),
	}
}

func NewConfigTypeByFormatVersion(name string, fmt runtime.FormatVersion[Config], usages ...string) ConfigType {
	return &configType{
		VersionedTypedObjectType: runtime.NewVersionedTypedObjectTypeByFormatVersion[Config](name, fmt),
		usage:                    strings.Join(usages, "\n"),
	}
}

func (t *configType) Usage() string {
	return t.usage
}
