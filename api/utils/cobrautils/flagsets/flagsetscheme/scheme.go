package flagsetscheme

import (
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
	"ocm.software/ocm/api/utils/logging"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/api/utils/runtime/descriptivetype"
)

////////////////////////////////////////////////////////////////////////////////

// ExtendedProviderTypeScheme is the appropriately extended scheme interface based on
// runtime.TypeScheme supporting an extended config provider interface.
type ExtendedProviderTypeScheme[T runtime.TypedObject, R TypedObjectType[T], P flagsets.ConfigTypeOptionSetConfigProvider] interface {
	descriptivetype.TypeScheme[T, R]

	CreateConfigTypeSetConfigProvider() P

	Unwrap() TypeScheme[T, R]
}

type _TypeScheme[T runtime.TypedObject, R TypedObjectType[T]] interface {
	TypeScheme[T, R]
}

type typeSchemeWrapper[T runtime.TypedObject, R TypedObjectType[T], P flagsets.ConfigTypeOptionSetConfigProvider] struct {
	_TypeScheme[T, R]
}

func (s *typeSchemeWrapper[T, R, P]) CreateConfigTypeSetConfigProvider() P {
	return s._TypeScheme.CreateConfigTypeSetConfigProvider().(P)
}

func (s *typeSchemeWrapper[T, R, P]) Unwrap() TypeScheme[T, R] {
	return s._TypeScheme
}

// NewVersionedTypeSchemeWrapper wraps a [VersionedTypeScheme] into a scheme returning a specialized config provider
// by casting the result. The type scheme constructor provides different implementations based on its
// arguments. This method here can be used to provide a type scheme returning the correct type.
func NewVersionedTypeSchemeWrapper[T runtime.VersionedTypedObject, R VersionTypedObjectType[T], P flagsets.ConfigTypeOptionSetConfigProvider](s VersionedTypeScheme[T, R]) ExtendedProviderTypeScheme[T, R, P] {
	return &typeSchemeWrapper[T, R, P]{s}
}

// NewTypeSchemeWrapper wraps a [TypeScheme] into a scheme returning a specialized config provider
// by casting the result. The type scheme constructor provides different implementations based on its
// arguments. This method here can be used to provide a type scheme returning the correct type.
func NewTypeSchemeWrapper[T runtime.TypedObject, R TypedObjectType[T], P flagsets.ConfigTypeOptionSetConfigProvider](s TypeScheme[T, R]) ExtendedProviderTypeScheme[T, R, P] {
	return &typeSchemeWrapper[T, R, P]{s}
}

// VersionedTypeScheme is the appropriately extended scheme interface based on
// runtime.TypeScheme.
type VersionedTypeScheme[T runtime.VersionedTypedObject, R VersionTypedObjectType[T]] interface {
	ExtendedProviderTypeScheme[T, R, flagsets.ConfigTypeOptionSetConfigProvider]
}

type TypeScheme[T runtime.TypedObject, R TypedObjectType[T]] interface {
	ExtendedProviderTypeScheme[T, R, flagsets.ConfigTypeOptionSetConfigProvider]
}

type _typeScheme[T runtime.TypedObject, R TypedObjectType[T]] interface {
	descriptivetype.TypeScheme[T, R]
}

type typeScheme[T runtime.TypedObject, R TypedObjectType[T], S TypeScheme[T, R]] struct {
	cfgname     string
	description string
	group       string
	typeOption  string
	_typeScheme[T, R]
}

func flagExtender[T runtime.TypedObject, R TypedObjectType[T]](ty R) string {
	if h := ty.ConfigOptionTypeSetHandler(); h != nil {
		return utils.IndentLines(flagsets.FormatConfigOptions(h), "  ")
	}
	return ""
}

// NewVersionedTypeScheme provides an VersionedTypeScheme implementation based on the interfaces
// and the default runtime.TypeScheme implementation.
func NewVersionedTypeScheme[T runtime.VersionedTypedObject, R VersionTypedObjectType[T], S TypeScheme[T, R]](kindname string, cfgname, typeOption, desc, group string, unknown runtime.Unstructured, acceptUnknown bool, base ...S) VersionedTypeScheme[T, R] {
	scheme := descriptivetype.NewVersionedTypeScheme[T, R](kindname, flagExtender[T, R], unknown, acceptUnknown, utils.Optional(base...))
	return &typeScheme[T, R, S]{
		cfgname:     cfgname,
		description: desc,
		group:       group,
		typeOption:  typeOption,
		_typeScheme: scheme,
	}
}

// NewTypeScheme provides an VersionedTypeScheme implementation based on the interfaces
// and the default runtime.TypeScheme implementation.
func NewTypeScheme[T runtime.TypedObject, R TypedObjectType[T], S TypeScheme[T, R]](kindname string, cfgname, typeOption, desc, group string, unknown runtime.Unstructured, acceptUnknown bool, base ...S) TypeScheme[T, R] {
	scheme := descriptivetype.NewTypeScheme[T, R](kindname, flagExtender[T, R], unknown, acceptUnknown, utils.Optional(base...))
	return &typeScheme[T, R, S]{
		cfgname:     cfgname,
		description: desc,
		group:       group,
		typeOption:  typeOption,
		_typeScheme: scheme,
	}
}

func (s *typeScheme[T, R, S]) Unwrap() TypeScheme[T, R] {
	return s
}

func (t *typeScheme[T, R, S]) CreateConfigTypeSetConfigProvider() flagsets.ConfigTypeOptionSetConfigProvider {
	var prov flagsets.ConfigTypeOptionSetConfigProvider
	if t.typeOption == "" {
		prov = flagsets.NewExplicitlyTypedConfigProvider(t.cfgname, t.description, true)
	} else {
		prov = flagsets.NewTypedConfigProvider(t.cfgname, t.description, t.typeOption, true)
	}
	if t.group != "" {
		prov.AddGroups(t.group)
	}
	for _, p := range t.KnownTypes() {
		err := prov.AddTypeSet(p.ConfigOptionTypeSetHandler())
		if err != nil {
			logging.Logger().LogError(err, "cannot compose type CLI options", "type", t.cfgname)
		}
	}
	if t.BaseScheme() != nil {
		base := t.BaseScheme().(S)
		for _, s := range base.CreateConfigTypeSetConfigProvider().OptionTypeSets() {
			if prov.GetTypeSet(s.GetName()) == nil {
				err := prov.AddTypeSet(s)
				if err != nil {
					logging.Logger().LogError(err, "cannot compose type CLI options", "type", t.cfgname)
				}
			}
		}
	}

	return prov
}

func (t *typeScheme[T, R, S]) KnownTypes() runtime.KnownTypes[T, R] {
	return t._typeScheme.KnownTypes() // Goland
}
