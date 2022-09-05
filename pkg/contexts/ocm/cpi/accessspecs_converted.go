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

package cpi

import (
	"encoding/json"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/core"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type AccessSpecConverter interface {
	ConvertFrom(object core.AccessSpec) (runtime.TypedObject, error)
	ConvertTo(object interface{}) (core.AccessSpec, error)
}

type AccessSpecVersion interface {
	AccessSpecConverter
	runtime.TypedObjectDecoder
	CreateData() interface{}
}

type accessSpecVersion struct {
	*runtime.ConvertingDecoder
	AccessSpecConverter
}

type typedObjectConverter struct {
	converter AccessSpecConverter
}

func (c *typedObjectConverter) ConvertTo(object interface{}) (runtime.TypedObject, error) {
	return c.converter.ConvertTo(object)
}

func NewAccessSpecVersion(proto runtime.TypedObject, converter AccessSpecConverter) AccessSpecVersion {
	return &accessSpecVersion{
		ConvertingDecoder:   runtime.MustNewConvertingDecoder(proto, &typedObjectConverter{converter}),
		AccessSpecConverter: converter,
	}
}

////////////////////////////////////////////////////////////////////////////////

type ConvertedAccessType struct {
	AccessSpecVersion
	accessType
}

var (
	_ AccessSpecVersion          = &ConvertedAccessType{}
	_ runtime.TypedObjectEncoder = &ConvertedAccessType{}
)

func NewConvertedAccessSpecType(name string, v AccessSpecVersion) *ConvertedAccessType {
	return &ConvertedAccessType{
		accessType: accessType{
			ObjectVersionedType: runtime.NewVersionedObjectType(name),
			TypedObjectDecoder:  v,
		},
		AccessSpecVersion: v,
	}
}

func (t *ConvertedAccessType) Encode(obj runtime.TypedObject, m runtime.Marshaler) ([]byte, error) {
	c, err := t.ConvertFrom(obj.(AccessSpec))
	if err != nil {
		return nil, err
	}
	return m.Marshal(c)
}

////////////////////////////////////////////////////////////////////////////////

func MarshalConvertedAccessSpec(ctx Context, s AccessSpec) ([]byte, error) {
	t := ctx.AccessMethods().GetAccessType(s.GetType())
	// fmt.Printf("found spec type %s: %T\n", s.Algorithm(), t)
	if c, ok := t.(AccessSpecConverter); ok {
		out, err := c.ConvertFrom(s)
		if err != nil {
			return nil, err
		}
		return json.Marshal(out)
	}
	return nil, errors.ErrNotImplemented("converted access version type", s.GetType(), s.GetKind())
}
