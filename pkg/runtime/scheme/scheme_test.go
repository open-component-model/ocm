// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package scheme_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/runtime/scheme"
	. "github.com/open-component-model/ocm/pkg/testutils"
)

type Object interface {
	runtime.VersionedTypedObject
	Selector() string
}

type Type interface {
	scheme.Type[Object]
}

////////////////////////////////////////////////////////////////////////////////
// KIND: version

const KIND_VERSION = "version"

type k1 struct {
	runtime.ObjectVersionedType
	Key string `json:"key"`

	Kind    string `json:"kind"`
	Version string `json:"version"`
}

func (o *k1) Selector() string {
	return o.Key
}

type k1v1 = k1

const K1V1Data = `
  type: version/v1
  key: k1v1-key
  kind: test
  version: v1.1.1
`

const K1V1ImplicitData = `
  type: version
  key: k1v1-key
  kind: test
  version: v1.1.1
`

var K1V1 = &k1v1{
	ObjectVersionedType: runtime.ObjectVersionedType{"version/v1"},
	Key:                 "k1v1-key",
	Kind:                "test",
	Version:             "v1.1.1",
}

type k1v2 = struct {
	runtime.ObjectVersionedType
	Key        string `json:"key"`
	APIVersion string `json:"apiVersion"`
}

const K1V2Data = `
  type: version/v2
  key: k1v1-key
  apiVersion: test/v1.1.1
`

type k1v2Converter struct{}

var _ scheme.Converter[Object] = (*k1v2Converter)(nil)

func (k k1v2Converter) ConvertTo(object interface{}) (Object, error) {
	in := object.(*k1v2)
	r := &k1{
		ObjectVersionedType: runtime.ObjectVersionedType{in.GetType()},
		Key:                 in.Key,
	}
	r.Kind, r.Version = runtime.KindVersion(in.APIVersion)
	return r, nil
}

func (k k1v2Converter) ConvertFrom(o Object) (runtime.TypedObject, error) {
	in := o.(*k1)
	r := &k1v2{
		ObjectVersionedType: runtime.ObjectVersionedType{o.GetType()},
		Key:                 in.Key,
		APIVersion:          runtime.TypeName(in.Kind, in.Version),
	}
	return r, nil
}

var _ = Describe("scheme", func() {
	var s scheme.Scheme[Object, Type]

	BeforeEach(func() {
		s = scheme.NewScheme[Object, Type]()
	})

	It("handles single version", func() {
		MustBeSuccessful(s.RegisterType(runtime.TypeName(KIND_VERSION, "v1"), scheme.NewIdentityType[Object](&k1v1{})))

		o := Must(s.Decode([]byte(K1V1Data), nil))
		Expect(o).To(Equal(K1V1))
	})

	It("handles implicit version", func() {
		MustBeSuccessful(s.RegisterType(runtime.TypeName(KIND_VERSION, "v1"), scheme.NewIdentityType[Object](&k1v1{})))

		o := Must(s.Decode([]byte(K1V1ImplicitData), nil))
		r := *K1V1
		r.Type = "version"
		Expect(o).To(Equal(&r))
	})

	It("handles converted version", func() {
		MustBeSuccessful(s.RegisterType(runtime.TypeName(KIND_VERSION, "v1"), scheme.NewIdentityType[Object](&k1v1{})))
		MustBeSuccessful(s.RegisterType(runtime.TypeName(KIND_VERSION, "v2"), scheme.NewTypeByProtoType[Object](&k1v2{}, k1v2Converter{})))

		o := Must(s.Decode([]byte(K1V2Data), nil))

		r := *K1V1
		r.Type = "version/v2"
		Expect(o).To(Equal(&r))
	})

	It("get versions", func() {
		MustBeSuccessful(s.RegisterType(runtime.TypeName(KIND_VERSION, "v1"), scheme.NewIdentityType[Object](&k1v1{})))
		MustBeSuccessful(s.RegisterType(runtime.TypeName(KIND_VERSION, "v2"), scheme.NewTypeByProtoType[Object](&k1v2{}, k1v2Converter{})))

		Expect(s.KnownKinds()).To(Equal([]string{"version"}))
		Expect(s.KnownVersions("version")).To(Equal([]string{"v1", "v2"}))

		Expect(len(s.KnownTypes())).To(Equal(3))
	})
})
