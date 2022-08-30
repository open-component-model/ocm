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
	//"github.com/open-component-model/ocm/pkg/json"
	"encoding/json"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/runtime"
)

var _ = Describe("json", func() {

	Context("what the fuck", func() {
		It("derived marshal", func() {
			d := &DerivedMarshal{
				Field: "value",
			}
			data, err := json.Marshal(d)
			Expect(err).To(Succeed())
			Expect(string(data)).To(Equal("{\"Test\":\"test\"}")) // !!!! Field is not marshalled
		})

		It("derived fake marshal", func() {
			var marshaler json.Marshaler

			d := &DerivedFakeMarshal{
				Field: "value",
			}

			ok := reflect.TypeOf(d).Implements(reflect.TypeOf(&marshaler).Elem())
			Expect(ok).To(BeFalse())

			data, err := json.Marshal(d)
			Expect(err).To(Succeed())
			Expect(string(data)).To(Equal("{\"Field\":\"value\"}")) // !!!! Base type not marshaled at all
		})

		It("using hidden marshal", func() {
			var marshaler json.Marshaler

			d := &UsingHidden{
				DerivedFakeMarshal: DerivedFakeMarshal{
					Marshal: Marshal{},
					Field:   "value",
				},
				Using: "using",
			}

			ok := reflect.TypeOf(d).Implements(reflect.TypeOf(&marshaler).Elem())
			Expect(ok).To(BeFalse())

			data, err := json.Marshal(d)
			Expect(err).To(Succeed())
			Expect(string(data)).To(Equal("{\"Field\":\"value\",\"Using\":\"using\"}")) // !!!! Base type not marshaled at all
		})

		It("wrapping unstructured map", func() {
			d := &WrappingUnstructured{
				Unstructured: Unstructured{
					runtime.ATTR_TYPE: "test",
				},
				Field: "value",
			}
			data, err := json.Marshal(d)
			Expect(err).To(Succeed())
			Expect(string(data)).To(Equal("{\"Unstructured\":{\"type\":\"test\"},\"Field\":\"value\"}")) // not inlined, why ?
		})

		It("wrapping unstructured map with inline tag", func() {
			d := &WrappingUnstructuredInlined{
				Unstructured: Unstructured{
					runtime.ATTR_TYPE: "test",
				},
				Field: "value",
			}
			data, err := json.Marshal(d)
			Expect(err).To(Succeed())
			Expect(string(data)).To(Equal("{\"Unstructured\":{\"type\":\"test\"},\"Field\":\"value\"}")) // not inlined, why
		})

		// inline of named field does not work
		It("non anonymous", func() {
			un := &NonAnonymous{
				ObjectType: runtime.ObjectType{"test"},
				X:          "value",
			}
			data, err := json.Marshal(un)
			Expect(err).To(Succeed())
			Expect(string(data)).To(Equal("{\"ObjectType\":{\"type\":\"test\"},\"TTT\":\"value\"}")) // !!!!!
		})
		It("anonymous", func() {
			un := &Anonymous{
				ObjectType: runtime.ObjectType{"test"},
				X:          "value",
			}
			data, err := json.Marshal(un)
			Expect(err).To(Succeed())
			Expect(string(data)).To(Equal("{\"type\":\"test\",\"TTT\":\"value\"}")) // !!!!!
		})
	})
})
