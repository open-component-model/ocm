// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package runtime_test

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	runtime "github.com/open-component-model/ocm/pkg/runtime2"
	. "github.com/open-component-model/ocm/pkg/testutils"
)

const (
	Type   = "testType"
	TypeV1 = Type + "/v1"
)

var versions runtime.Scheme[TestSpec]

func init() {
	var ts TestSpec
	versions = runtime.MustNewDefaultScheme[TestSpec](&ts, nil, false, nil)

	versions.RegisterByDecoder(Type, runtime.NewVersionedTypedObjectTypeByConverter[TestSpec, *TestSpec1](Type, &Spec1V1{}, &converterSpec1V1{}))
	versions.RegisterByDecoder(TypeV1, runtime.NewVersionedTypedObjectTypeByConverter[TestSpec, *TestSpec1](TypeV1, &Spec1V1{}, &converterSpec1V1{}))
}

type TestSpec interface {
	runtime.VersionedTypedObject
	TestFunction()
}

type TestSpec1 struct {
	runtime.InternalVersionedTypedObject[TestSpec]
	Field string `json:"field"`
}

func (a TestSpec1) MarshalJSON() ([]byte, error) {
	return runtime.MarshalVersionedTypedObject(&a)
}

func (a *TestSpec1) TestFunction() {}

func NewTestSpec1(field string) *TestSpec1 {
	return &TestSpec1{
		InternalVersionedTypedObject: runtime.NewInternalVersionedTypedObject[TestSpec](versions, Type),
		Field:                        field,
	}
}

type Spec1V1 struct {
	runtime.ObjectVersionedType
	OldField string `json:"oldField"`
}

type converterSpec1V1 struct{}

var _ runtime.Converter[*TestSpec1] = (*converterSpec1V1)(nil)

func (_ converterSpec1V1) ConvertFrom(in *TestSpec1) (runtime.TypedObject, error) {
	return &Spec1V1{
		ObjectVersionedType: runtime.NewVersionedObjectType(in.Type),
		OldField:            in.Field,
	}, nil
}

func (_ converterSpec1V1) ConvertTo(object interface{}) (*TestSpec1, error) {
	in, ok := object.(*Spec1V1)
	if !ok {
		return nil, fmt.Errorf("failed to assert type %T to Spec1V1", object)
	}
	return &TestSpec1{
		InternalVersionedTypedObject: runtime.NewInternalVersionedTypedObject[TestSpec](versions, in.Type),
		Field:                        in.OldField,
	}, nil
}

type encoder interface {
	getEncoder() int
}

type object struct {
}

func (_ *object) getEncoder() int {
	return 1
}

var _ = Describe("versioned types", func() {
	It("marshals version", func() {
		s1 := NewTestSpec1("value")

		data := Must(json.Marshal(s1))
		Expect(string(data)).To(StringEqualWithContext(`{"type":"testType","oldField":"value"}`))

		spec := Must(versions.Decode(data, runtime.DefaultJSONEncoding))
		Expect(spec).To(Equal(s1))
	})
})
