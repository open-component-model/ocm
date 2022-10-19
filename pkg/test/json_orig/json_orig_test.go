// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

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
