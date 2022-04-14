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

package runtime_test

import (
	"encoding/json"
	"fmt"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/runtime"
)

func InOut(in runtime.TypedObject, encoding runtime.Encoding) (runtime.TypedObject, string, error) {
	t := reflect.TypeOf(in)
	fmt.Printf("in: %s\n", t)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var p reflect.Value

	if t.Kind() == reflect.Map {
		p = reflect.New(t)
		m := reflect.MakeMap(t)
		fmt.Printf("pointer: %s\n", p.Type())
		p.Elem().Set(m)
	} else {
		p = reflect.New(t)
	}
	out := p.Interface().(runtime.TypedObject)

	fmt.Printf("out: %T\n", out)
	data, err := encoding.Marshal(in)
	if err != nil {
		return nil, "", err
	}
	err = encoding.Unmarshal(data, out)
	return out.(runtime.TypedObject), string(data), err
}

var _ = Describe("*** unstructured", func() {
	result := "{\"type\":\"test\"}"

	It("unmarshal simple unstructured", func() {
		un := runtime.NewEmptyUnstructured("test")
		data, err := json.Marshal(un)
		Expect(err).To(Succeed())
		Expect(string(data)).To(Equal("{\"type\":\"test\"}"))

		un = &runtime.UnstructuredTypedObject{}
		fmt.Printf("out: %T\n", un)
		err = json.Unmarshal(data, un)
		Expect(err).To(Succeed())
		Expect(un.GetType()).To(Equal("test"))
	})

	It("unmarshal json test", func() {
		out, data, err := InOut(runtime.NewEmptyUnstructured("test"), runtime.DefaultJSONEncoding)
		Expect(err).To(Succeed())
		Expect(out.GetType()).To(Equal("test"))
		Expect(data).To(Equal(result))

		out, data, err = InOut(runtime.NewEmptyUnstructuredVersioned("test"), runtime.DefaultJSONEncoding)
		Expect(err).To(Succeed())
		Expect(out.GetType()).To(Equal("test"))
		Expect(data).To(Equal(result))
	})

	It("unmarshal yaml test", func() {
		out, data, err := InOut(runtime.NewEmptyUnstructured("test"), runtime.DefaultYAMLEncoding)
		Expect(err).To(Succeed())
		Expect(out.GetType()).To(Equal("test"))
		Expect(data).To(Equal("type: test\n"))

		out, data, err = InOut(runtime.NewEmptyUnstructuredVersioned("test"), runtime.DefaultYAMLEncoding)
		Expect(err).To(Succeed())
		Expect(out.GetType()).To(Equal("test"))
		Expect(data).To(Equal("type: test\n"))
	})
})
