package ppi

import (
	"encoding/json"
	"reflect"
	"slices"

	"github.com/pkg/errors"

	"ocm.software/ocm/api/ocm/extensions/accessmethods/options"
	"ocm.software/ocm/api/utils/runtime"
)

type decoder runtime.TypedObjectDecoder[runtime.TypedObject]

type AccessMethodBase struct {
	decoder
	nameDescription

	version string
	format  string
}

func MustNewAccessMethodBase(name, version string, proto AccessSpec, desc string, format string) AccessMethodBase {
	decoder, err := runtime.NewDirectDecoder(proto)
	if err != nil {
		panic(err)
	}

	return AccessMethodBase{
		decoder: decoder,
		nameDescription: nameDescription{
			name: name,
			desc: desc,
		},
		version: version,
		format:  format,
	}
}

func (b *AccessMethodBase) Version() string {
	return b.version
}

func (b *AccessMethodBase) Format() string {
	return b.format
}

////////////////////////////////////////////////////////////////////////////////

type UploaderBase = nameDescription

func MustNewUploaderBase(name, desc string) UploaderBase {
	return UploaderBase{
		name: name,
		desc: desc,
	}
}

////////////////////////////////////////////////////////////////////////////////

type ValueSetBase struct {
	decoder
	nameDescription

	version string
	format  string

	purposes []string
}

func MustNewValueSetBase(name, version string, proto runtime.TypedObject, purposes []string, desc string, format string) ValueSetBase {
	decoder, err := runtime.NewDirectDecoder(proto)
	if err != nil {
		panic(err)
	}
	return ValueSetBase{
		decoder: decoder,
		nameDescription: nameDescription{
			name: name,
			desc: desc,
		},
		version:  version,
		format:   format,
		purposes: slices.Clone(purposes),
	}
}

func (b *ValueSetBase) Version() string {
	return b.version
}

func (b *ValueSetBase) Format() string {
	return b.format
}

func (b *ValueSetBase) Purposes() []string {
	return b.purposes
}

////////////////////////////////////////////////////////////////////////////////

type nameDescription struct {
	name string
	desc string
}

func (b *nameDescription) Name() string {
	return b.name
}

func (b *nameDescription) Description() string {
	return b.desc
}

////////////////////////////////////////////////////////////////////////////////

// Config is a generic structured config stored in a string map.
type Config map[string]interface{}

func (c Config) GetValue(name string) (interface{}, bool) {
	v, ok := c[name]
	return v, ok
}

func (c Config) ConvertFor(list ...options.OptionType) error {
	for _, o := range list {
		if v, ok := c[o.GetName()]; ok {
			t := reflect.TypeOf(o.Create().Value())
			if t != reflect.TypeOf(v) {
				data, err := json.Marshal(v)
				if err != nil {
					return errors.Wrapf(err, "cannot marshal option value for %q", o.GetName())
				}
				value := reflect.New(t)
				err = json.Unmarshal(data, value.Interface())
				if err != nil {
					return errors.Wrapf(err, "cannot unmarshal option value for %q[%s]", o.GetName(), o.ValueType())
				}
				c[o.GetName()] = value.Elem().Interface()
			}
		}
	}
	return nil
}
