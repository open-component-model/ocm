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

package vfsattr

import (
	"fmt"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const ATTR_KEY = "github.com/mandelsoft/vfs"
const ATTR_SHORT = "vfs"

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
*intern* (not via command line)
Virtual filesystem to use for command line context.
`
}

func (a AttributeType) Encode(v interface{}, marshaller runtime.Marshaler) ([]byte, error) {
	if _, ok := v.(vfs.FileSystem); !ok {
		return nil, fmt.Errorf("vfs.FileSystem required")
	}
	return nil, nil
}

func (a AttributeType) Decode(data []byte, unmarshaller runtime.Unmarshaler) (interface{}, error) {
	return nil, errors.ErrNotSupported("decode attribute", ATTR_KEY)
}

////////////////////////////////////////////////////////////////////////////////

var _osfs = osfs.New()

func Get(ctx datacontext.Context) vfs.FileSystem {
	v := ctx.GetAttributes().GetAttribute(ATTR_KEY)
	if v == nil {
		return _osfs
	}
	fs, _ := v.(vfs.FileSystem)
	return fs

}

func Set(ctx datacontext.Context, fs vfs.FileSystem) {
	ctx.GetAttributes().SetAttribute(ATTR_KEY, fs)
}
