package internal

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/goutils/maputils"
	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/modern-go/reflect2"

	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
	"ocm.software/ocm/api/utils/cobrautils/flagsets/flagsetscheme"
	"ocm.software/ocm/api/utils/runtime"
)

type Context = cpi.Context

type EntryType flagsetscheme.VersionTypedObjectType[Entry]

// Entry is the interface access method specifications
// must fulfill. The main task is to map the specification
// to a concrete implementation of the access method for a dedicated
// component version.
type Entry interface {
	runtime.VersionedTypedObject

	Describe(ctx Context) string
	Validate(ctx Context) error
}

type (
	EntryDecoder      = runtime.TypedObjectDecoder[Entry]
	EntryTypeProvider = runtime.KnownTypesProvider[Entry, EntryType]
)

////////////////////////////////////////////////////////////////////////////////

type EntryTypeScheme = flagsetscheme.ExtendedTypeScheme[Entry, EntryType, flagsets.ExplicitlyTypedConfigTypeOptionSetConfigProvider]

func unwrapTypeScheme(s EntryTypeScheme) flagsetscheme.TypeScheme[Entry, EntryType] {
	return s.Unwrap()
}

func NewEntryTypeScheme(base ...EntryTypeScheme) EntryTypeScheme {
	return flagsetscheme.NewTypeSchemeWrapper[Entry, EntryType, flagsets.ExplicitlyTypedConfigTypeOptionSetConfigProvider](flagsetscheme.NewTypeScheme[Entry, EntryType, flagsetscheme.TypeScheme[Entry, EntryType]]("Entry type", "entry", "", "routing slip entry specification", "Entry Specification Options", &UnknownEntry{}, true, sliceutils.Transform(base, unwrapTypeScheme)...))
}

func NewStrictEntryTypeScheme(base ...EntryTypeScheme) EntryTypeScheme {
	return flagsetscheme.NewTypeSchemeWrapper[Entry, EntryType, flagsets.ExplicitlyTypedConfigTypeOptionSetConfigProvider](flagsetscheme.NewTypeScheme[Entry, EntryType, flagsetscheme.TypeScheme[Entry, EntryType]]("Entry type", "entry", "", "routing slip entry specification", "Entry Specification Options", nil, false, sliceutils.Transform(base, unwrapTypeScheme)...))
}

func CreateEntry(t runtime.TypedObject) (Entry, error) {
	return defaultEntryTypeScheme.Convert(t)
}

////////////////////////////////////////////////////////////////////////////////

type UnknownEntry struct {
	runtime.UnstructuredVersionedTypedObject `json:",inline"`
}

var (
	_ runtime.TypedObject = &UnknownEntry{}
	_ runtime.Unknown     = &UnknownEntry{}
)

func (_ *UnknownEntry) IsUnknown() bool {
	return true
}

func (u *UnknownEntry) Describe(ctx Context) string {
	keys := maputils.OrderedKeys(u.Object)
	cnt := 0
	desc := []string{}
	delta := 0
	for _, k := range keys {
		v := u.Object[k]
		if k == runtime.ATTR_TYPE {
			delta = 1
			continue
		}
		if v == nil {
			continue
		}
		if cnt > 3 {
			continue
		}
		value := reflect.ValueOf(v)
		if value.Kind() == reflect.Array || value.Kind() == reflect.Map || value.Kind() == reflect.Slice {
			continue
		}
		cnt++
		desc = append(desc, fmt.Sprintf("%s: %v", k, v))
	}
	if len(desc) == 0 {
		return "<unknown type>"
	}
	if len(keys) > len(desc)+delta {
		return strings.Join(desc, ", ") + ", ..."
	}
	return strings.Join(desc, ", ")
}

func (u *UnknownEntry) Validate(ctx Context) error {
	return nil
}

var _ Entry = &UnknownEntry{}

////////////////////////////////////////////////////////////////////////////////

type EvaluatableEntry interface {
	Entry
	Evaluate(ctx Context) (Entry, error)
}

type GenericEntry struct {
	runtime.UnstructuredVersionedTypedObject `json:",inline"`
}

var _ Entry = &GenericEntry{}

func AsGenericEntry(u *runtime.UnstructuredTypedObject) *GenericEntry {
	return &GenericEntry{runtime.UnstructuredVersionedTypedObject{*u}}
}

func ToGenericEntry(spec Entry) (*GenericEntry, error) {
	if reflect2.IsNil(spec) {
		return nil, nil
	}
	if g, ok := spec.(*GenericEntry); ok {
		return g, nil
	}
	data, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}
	return newGenericEntry(data, runtime.DefaultJSONEncoding)
}

func NewGenericEntry(data []byte, unmarshaler ...runtime.Unmarshaler) (Entry, error) {
	return generics.CastPointerR[Entry](newGenericEntry(data, utils.Optional(unmarshaler...)))
}

func newGenericEntry(data []byte, unmarshaler runtime.Unmarshaler) (*GenericEntry, error) {
	unstr := &runtime.UnstructuredVersionedTypedObject{}
	if unmarshaler == nil {
		unmarshaler = runtime.DefaultYAMLEncoding
	}
	err := unmarshaler.Unmarshal(data, unstr)
	if err != nil {
		return nil, err
	}
	return &GenericEntry{*unstr}, nil
}

func (s *GenericEntry) Describe(ctx Context) string {
	eff, err := s.Evaluate(ctx)
	if err != nil {
		return fmt.Sprintf("invalid access specification: %s", err.Error())
	}
	return eff.Describe(ctx)
}

func (s *GenericEntry) Validate(ctx Context) error {
	eff, err := s.Evaluate(ctx)
	if err != nil {
		return errors.Wrapf(err, "invalid access specification")
	}
	if _, ok := eff.(*GenericEntry); ok {
		return nil
	}
	return eff.Validate(ctx)
}

func (s *GenericEntry) Evaluate(ctx Context) (Entry, error) {
	raw, err := s.GetRaw()
	if err != nil {
		return nil, err
	}
	return For(ctx).Decode(raw, runtime.DefaultJSONEncoding) // TODO: switch to context
}

// defaultEntryTypeScheme contains all globally known access serializer.
var defaultEntryTypeScheme = NewEntryTypeScheme()

func DefaultEntryTypeScheme() EntryTypeScheme {
	return defaultEntryTypeScheme
}
