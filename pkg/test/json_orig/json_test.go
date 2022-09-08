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

package test_test

import (
	"encoding/json"
	"reflect"

	"github.com/sirupsen/logrus"

	"github.com/open-component-model/ocm/pkg/runtime"
)

func InOut(in runtime.TypedObject, encoding runtime.Encoding) (runtime.TypedObject, string, error) {
	t := reflect.TypeOf(in)
	logrus.Infof("in: %s\n", t)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var p reflect.Value

	if t.Kind() == reflect.Map {
		p = reflect.New(t)
		m := reflect.MakeMap(t)
		logrus.Infof("pointer: %s\n", p.Type())
		p.Elem().Set(m)
	} else {
		p = reflect.New(t)
	}
	out := p.Interface().(runtime.TypedObject)

	logrus.Infof("out: %T\n", out)
	data, err := encoding.Marshal(in)
	if err != nil {
		return nil, "", err
	}
	err = encoding.Unmarshal(data, out)
	return out, string(data), err
}

type NonAnonymous struct {
	ObjectType runtime.ObjectType `json:",inline"`
	X          string             `json:"TTT"`
}

type Anonymous struct {
	runtime.ObjectType `json:",inline"`
	X                  string `json:"TTT"`
}

type Marshal struct{}

func (m *Marshal) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Test string
	}{
		Test: "test",
	})
}

type DerivedMarshal struct {
	Marshal
	Field string
}

var _ json.Marshaler = &DerivedMarshal{}

type DerivedFakeMarshal struct {
	Marshal
	Field string
}

func (m *DerivedFakeMarshal) MarshalJSON() {}

type UsingHidden struct {
	DerivedFakeMarshal
	Using string
}

type Unstructured map[string]interface{}

type WrappingUnstructured struct {
	Unstructured
	Field string
}

type WrappingUnstructuredInlined struct {
	Unstructured `json:",inline"`
	Field        string
}
