// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package runtime

// github.com/ghodss/yaml

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
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

var DefaultJSONEncoding = &EncodingWrapper{
	Marshaler:   MarshalFunction(json.Marshal),
	Unmarshaler: UnmarshalFunction(json.Unmarshal),
}

var DefaultYAMLEncoding = &EncodingWrapper{
	Marshaler:   MarshalFunction(yaml.Marshal),
	Unmarshaler: UnmarshalFunction(yaml.Unmarshal),
}
