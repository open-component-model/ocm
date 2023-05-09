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

	versions.RegisterByDecoder(Type, runtime.NewConvertedType[TestSpec](Type, vSpec1V1))
	versions.RegisterByDecoder(TypeV1, runtime.NewConvertedType[TestSpec](TypeV1, vSpec1V1))
}

type TestSpec interface {
	runtime.VersionedTypedObject
}

type TestSpec1 struct {
	runtime.InternalVersionedObjectType[TestSpec]
	Field string `json:"field"`
}

func (a TestSpec1) MarshalJSON() ([]byte, error) {
	return runtime.MarshalObjectVersionedType(&a)
}

func NewTestSpec1(field string) *TestSpec1 {
	return &TestSpec1{
		InternalVersionedObjectType: runtime.NewInternalVersionedObjectType[TestSpec](versions, Type),
		Field:                       field,
	}
}

var vSpec1V1 = runtime.NewProtoBasedVersion[TestSpec](&Spec1V1{}, converterSpec1V1{})

type Spec1V1 struct {
	runtime.ObjectVersionedType
	OldField string `json:"oldField"`
}

type converterSpec1V1 struct{}

func (_ converterSpec1V1) ConvertFrom(object TestSpec) (runtime.TypedObject, error) {
	in, ok := object.(*TestSpec1)
	if !ok {
		return nil, fmt.Errorf("failed to assert type %T to Spec1V1", object)
	}
	return &Spec1V1{
		ObjectVersionedType: runtime.NewVersionedObjectType(in.Type),
		OldField:            in.Field,
	}, nil
}

func (_ converterSpec1V1) ConvertTo(object interface{}) (TestSpec, error) {
	in, ok := object.(*Spec1V1)
	if !ok {
		return nil, fmt.Errorf("failed to assert type %T to AccessSpecV1", object)
	}
	return &TestSpec1{
		InternalVersionedObjectType: runtime.NewInternalVersionedObjectType[TestSpec](versions, in.Type),
		Field:                       in.OldField,
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
	})
})
