// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

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
