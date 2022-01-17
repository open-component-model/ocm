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

package accesstypes

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/ocm/core"
	"github.com/gardener/ocm/pkg/ocm/runtime"
)

type AccessSpecConverter interface {
	ConvertFrom(object core.AccessSpec) (runtime.TypedObject, error)
	ConvertTo(object runtime.TypedObject) (core.AccessSpec, error)
}

type AccessSpecVersion interface {
	AccessSpecConverter
	CreateData() runtime.TypedObject
}

type accessSpecVersion struct {
	AccessSpecConverter
	spectype reflect.Type
}

func NewAccessSpecVersion(proto runtime.TypedObject, converter AccessSpecConverter) AccessSpecVersion {
	return &accessSpecVersion{
		spectype:            ProtoType(proto),
		AccessSpecConverter: converter,
	}
}

func (v *accessSpecVersion) CreateData() runtime.TypedObject {
	return reflect.New(v.spectype).Interface().(runtime.TypedObject)
}

func (v *accessSpecVersion) Converter() AccessSpecConverter {
	return v.AccessSpecConverter
}

////////////////////////////////////////////////////////////////////////////////

type ConvertedAccessType struct {
	accessType
	AccessSpecConverter
}

func NewConvertedType(name string, v AccessSpecVersion) *ConvertedAccessType {
	return &ConvertedAccessType{
		accessType: accessType{
			factory: v.CreateData,
			name:    name,
		},
		AccessSpecConverter: v,
	}
}

func (t *ConvertedAccessType) Decode(data []byte) (runtime.TypedObject, error) {
	obj, err := t.accessType.Decode(data)
	if err != nil {
		return nil, err
	}
	return t.ConvertTo(obj)
}

////////////////////////////////////////////////////////////////////////////////

func MarshalConvertedAccessSpec(s core.AccessSpec) ([]byte, error) {
	t := core.GetAccessType(s.GetType())
	fmt.Printf("found spec type %s: %T\n", s.GetType(), t)
	if c, ok := t.(AccessSpecConverter); ok {
		out, err := c.ConvertFrom(s)
		if err != nil {
			return nil, err
		}
		return json.Marshal(out)
	}
	return nil, errors.ErrNotImplemented("converted access version type", s.GetType(), s.GetName())
}
