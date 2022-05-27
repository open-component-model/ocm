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

package signingattr

import (
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/signing"
)

const ATTR_KEY = "github.com/mandelsoft/ocm/signing"
const ATTR_SHORT = "signing"

func init() {
	datacontext.RegisterAttributeType(ATTR_KEY, AttributeType{})
}

type AttributeType struct {
}

func (a AttributeType) Name() string {
	return ATTR_KEY
}

func (a AttributeType) Description() string {
	return `
*bool*
Public and private Key settings.
<pre>
{
  "publicKeys"": [
     "<provider>": {
       "data": "<base64>"
     }
  ],
  "privateKeys"": [
     "<provider>": {
       "path": "<file path>"
     }
  ]
</pre>
`
}

func (a AttributeType) Encode(v interface{}, marshaller runtime.Marshaler) ([]byte, error) {
	return nil, errors.ErrNotSupported("encoding of key registry")
}

func (a AttributeType) Decode(data []byte, unmarshaller runtime.Unmarshaler) (interface{}, error) {
	var value ConfigSpec
	err := unmarshaller.Unmarshal(data, &value)
	if err != nil {
		return nil, err
	}
	value.SetType(ConfigType)
	registry := signing.NewRegistry(signing.DefaultHandlerRegistry(), signing.DefaultKeyRegistry())
	value.ApplyToRegistry(registry)
	return registry, err
}

////////////////////////////////////////////////////////////////////////////////

func Get(ctx datacontext.Context) signing.Registry {
	a := ctx.GetAttributes().GetAttribute(ATTR_KEY)
	if a == nil {
		return signing.DefaultRegistry()
	}
	return a.(signing.Registry)
}

func Set(ctx datacontext.Context, registry signing.KeyRegistry) error {
	if _, ok := registry.(signing.Registry); !ok {
		registry = signing.NewRegistry(signing.DefaultHandlerRegistry(), registry)
	}
	return ctx.GetAttributes().SetAttribute(ATTR_KEY, registry)
}
