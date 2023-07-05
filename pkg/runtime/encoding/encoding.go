// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package encoding

// github.com/ghodss/yaml

import (
	"bytes"
	"encoding/json"

	"sigs.k8s.io/yaml"
)

type Marshaler interface {
	Marshal(obj interface{}) ([]byte, error)
}

type Unmarshaler interface {
	Unmarshal(data []byte, obj interface{}) error
}

type MarshalFunction func(obj interface{}) ([]byte, error)

func (f MarshalFunction) Marshal(obj interface{}) ([]byte, error) { return f(obj) }

type UnmarshalFunction func(data []byte, obj interface{}) error

func (f UnmarshalFunction) Unmarshal(data []byte, obj interface{}) error { return f(data, obj) }

type Encoding interface {
	Unmarshaler
	Marshaler
}

type EncodingWrapper struct {
	Unmarshaler
	Marshaler
}

// Cannot use Strict basic encoding, because
// all places using an incremental field parsing will not work anymore.
var (
	DefaultYAMLEncoding = StandardYAMLEncoding
	DefaultJSONEncoding = StandardJSONEncoding
)

var (
	StandardJSONEncoding = &EncodingWrapper{
		Marshaler:   MarshalFunction(json.Marshal),
		Unmarshaler: UnmarshalFunction(json.Unmarshal),
	}
	StandardYAMLEncoding = &EncodingWrapper{
		Marshaler:   MarshalFunction(yaml.Marshal),
		Unmarshaler: UnmarshalFunction(func(data []byte, obj interface{}) error { return yaml.Unmarshal(data, obj) }),
	}
)

var (
	StrictJSONEncoding = &EncodingWrapper{
		Marshaler: MarshalFunction(json.Marshal),
		Unmarshaler: UnmarshalFunction(func(data []byte, obj interface{}) error {
			d := json.NewDecoder(bytes.NewReader(data))
			d.DisallowUnknownFields()
			return d.Decode(obj)
		}),
	}
	StrictYAMLEncoding = &EncodingWrapper{
		Marshaler:   MarshalFunction(yaml.Marshal),
		Unmarshaler: UnmarshalFunction(func(data []byte, obj interface{}) error { return yaml.Unmarshal(data, obj, yaml.DisallowUnknownFields) }),
	}
)
