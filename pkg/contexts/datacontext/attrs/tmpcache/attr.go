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

package tmpcache

import (
	"fmt"
	"os"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/vfsattr"
	"github.com/open-component-model/ocm/pkg/errors"

	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const ATTR_KEY = "github.com/mandelsoft/tempblobcache"
const ATTR_SHORT = "blobcache"

func init() {
	datacontext.RegisterAttributeType(ATTR_KEY, AttributeType{}, ATTR_SHORT)
}

type AttributeType struct {
}

func (a AttributeType) Name() string {
	return ATTR_KEY
}

func (a AttributeType) Description() string {
	return `
*string* Foldername for temporary blob cache
The temporary blob cache is used to accessing large blobs from remote sytems.
The are temporarily stored in the filesystem, instead of the memory, to avoid
blowing up the memory consumption.
`
}

func (a AttributeType) Encode(v interface{}, marshaller runtime.Marshaler) ([]byte, error) {
	if a, ok := v.(*Attribute); !ok {
		return nil, fmt.Errorf("temppcache attribute")
	} else {
		return []byte(a.Path), nil
	}
}

func (a AttributeType) Decode(data []byte, unmarshaller runtime.Unmarshaler) (interface{}, error) {
	var s string
	err := runtime.DefaultYAMLEncoding.Unmarshal(data, &s)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid atribute value for %s", ATTR_KEY)
	}
	return &Attribute{
		Path: s,
	}, nil
}

////////////////////////////////////////////////////////////////////////////////

type Attribute struct {
	Path       string
	Filesystem vfs.FileSystem
}

func (a *Attribute) CreateTempFile(pat string) (vfs.File, error) {
	err := a.Filesystem.MkdirAll(a.Path, 0777)
	if err != nil {
		return nil, err
	}
	return vfs.TempFile(a.Filesystem, a.Path, pat)
}

////////////////////////////////////////////////////////////////////////////////

var def = &Attribute{
	Path: os.TempDir(),
}

func Get(ctx datacontext.Context) *Attribute {
	v := ctx.GetAttributes().GetAttribute(ATTR_KEY)
	a := def

	if v != nil {
		a, _ = v.(*Attribute)
	}
	return &Attribute{a.Path, vfsattr.Get(ctx)}
}

func Set(ctx datacontext.Context, a *Attribute) {
	ctx.GetAttributes().SetAttribute(ATTR_KEY, a)
}
