// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/open-component-model/ocm/pkg/runtime/encoding"
)

// github.com/ghodss/yaml

type Marshaler = encoding.Marshaler

type Unmarshaler encoding.Unmarshaler

type MarshalFunction func(obj interface{}) ([]byte, error)

func (f MarshalFunction) Marshal(obj interface{}) ([]byte, error) { return f(obj) }

type UnmarshalFunction func(data []byte, obj interface{}) error

func (f UnmarshalFunction) Unmarshal(data []byte, obj interface{}) error { return f(data, obj) }

type Encoding = encoding.Encoding

type EncodingWrapper = encoding.EncodingWrapper

var DefaultJSONEncoding = encoding.DefaultJSONEncoding

var DefaultYAMLEncoding = encoding.DefaultYAMLEncoding
