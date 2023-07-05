// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package validation_test

import (
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/runtime/validation"
	. "github.com/open-component-model/ocm/pkg/testutils"
)

type TestType struct {
	Field1 string `json:"field1,omitempty"`
	Field2 string `json:"field2,omitempty"`

	Struct1 StructType `json:"struct1,omitempty"`
	Struct2 StructType `json:"struct2,omitempty"`

	List1 []StructType `json:"list1,omitempty"`
	List2 []StructType `json:"list2,omitempty"`
}

type StructType struct {
	Field1 string `json:"structField1,omitempty"`
	Field2 string `json:"structField2,omitempty"`
}

var _ = Describe("validation", func() {

	It("validates correct data", func() {
		data := `
field1: value1
struct1:
  structField1: value2
list1:
- structField1: value3
`
		o := Must(validation.UnmarshalWithValidation[TestType]([]byte(data), runtime.DefaultYAMLEncoding, validation.NoAdditionalProperties()))

		Expect(reflect.TypeOf(o)).To(Equal(reflect.TypeOf(&TestType{})))
	})

	It("complains about unused fields", func() {
		data := `
any: value1
struct1:
  other: value2
list1:
- value: value3
`
		_, err := validation.UnmarshalWithValidation[TestType]([]byte(data), runtime.DefaultYAMLEncoding, validation.NoAdditionalProperties())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("[any: Forbidden: unknown field, list1[0].value: Forbidden: unknown field, struct1.other: Forbidden: unknown field]"))
	})

	It("complains about wrong structure (struct)", func() {
		data := `
struct1:
- structField1: value1
`
		_, err := validation.UnmarshalWithValidation[TestType]([]byte(data), runtime.DefaultYAMLEncoding, validation.NoAdditionalProperties())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("error unmarshaling JSON: while decoding JSON: json: cannot unmarshal array into Go struct field TestType.struct1 of type validation_test.StructType"))
	})

	It("complains about wrong structure (list)", func() {
		data := `
list1:
  structField1: value1
`
		_, err := validation.UnmarshalWithValidation[TestType]([]byte(data), runtime.DefaultYAMLEncoding, validation.NoAdditionalProperties())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("error unmarshaling JSON: while decoding JSON: json: cannot unmarshal object into Go struct field TestType.list1 of type []validation_test.StructType"))
	})

	It("complains about unused fields, but not top level", func() {
		data := `
any: value1
struct1:
  other: value2
list1:
- value: value3
`
		_, err := validation.UnmarshalWithValidation[TestType]([]byte(data), runtime.DefaultYAMLEncoding, validation.AdditionalRootProperties())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("[list1[0].value: Forbidden: unknown field, struct1.other: Forbidden: unknown field]"))
	})

	It("complains about unused fields, but not struct1 level", func() {
		data := `
any: value1
struct1:
  other: value2
list1:
- value: value3
`
		_, err := validation.UnmarshalWithValidation[TestType]([]byte(data), runtime.DefaultYAMLEncoding, validation.MapField("struct1", validation.AdditionalMapField("other")))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("[any: Forbidden: unknown field, list1[0].value: Forbidden: unknown field]"))
	})
})
