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

package keepblobattr

import (
	"fmt"

	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	ATTR_KEY   = "github.com/mandelsoft/ocm/keeplocalblob"
	ATTR_SHORT = "keeplocalblob"
)

func init() {
	datacontext.RegisterAttributeType(ATTR_KEY, AttributeType{}, ATTR_SHORT)
}

type AttributeType struct{}

func (a AttributeType) Name() string {
	return ATTR_KEY
}

func (a AttributeType) Description() string {
	return `
*bool*
Keep local blobs when importing OCI artefacts to OCI registries from <code>localBlob</code>
access methods. By default they will be expanded to OCI artefacts with the
access method <code>ociRegistry</code>. If this option is set to true, they will be stored
as local blobs, also. The access method will still be <code>localBlob</code> but with a nested
<code>ociRegistry</code> access method for describing the global access.
`
}

func (a AttributeType) Encode(v interface{}, marshaller runtime.Marshaler) ([]byte, error) {
	if _, ok := v.(bool); !ok {
		return nil, fmt.Errorf("boolean required")
	}
	return marshaller.Marshal(v)
}

func (a AttributeType) Decode(data []byte, unmarshaller runtime.Unmarshaler) (interface{}, error) {
	var value bool
	err := unmarshaller.Unmarshal(data, &value)
	return value, err
}

////////////////////////////////////////////////////////////////////////////////

func Get(ctx datacontext.Context) bool {
	a := ctx.GetAttributes().GetAttribute(ATTR_KEY)
	if a == nil {
		return false
	}
	return a.(bool)
}

func Set(ctx datacontext.Context, flag bool) error {
	return ctx.GetAttributes().SetAttribute(ATTR_KEY, flag)
}
