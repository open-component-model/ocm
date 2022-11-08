// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package runtime_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/utils"
)

type T1 struct {
	runtime.ObjectType
	T1 string
}

type T2 struct {
	runtime.ObjectType
	T2 string
}

var T1Decoder = runtime.MustNewDirectDecoder(&T1{})
var T2Decoder = runtime.MustNewDirectDecoder(&T2{})

var t1data = []byte(`{"type":"t1","t1":"v1"}`)
var t2data = []byte(`{"type":"t2","t2":"v2"}`)

var t1 = &T1{runtime.ObjectType{"t1"}, "v1"}
var t2 = &T2{runtime.ObjectType{"t2"}, "v2"}

var _ = Describe("scheme", func() {
	var scheme runtime.Scheme

	BeforeEach(func() {
		var rt runtime.TypedObject
		scheme = Must(runtime.NewDefaultScheme(&rt, &runtime.UnstructuredTypedObject{}, false, nil))
		MustBeSuccessful(scheme.RegisterByDecoder("t1", T1Decoder))
	})

	It("decodes object", func() {
		Expect(Must(scheme.Decode(t1data, nil))).To(Equal(t1))
		Expect(scheme.KnownTypeNames()).To(Equal([]string{"t1"}))
		Expect(utils.StringMapKeys(scheme.KnownTypes())).To(Equal([]string{"t1"}))
	})

	It("handles derived scheme", func() {
		var rt runtime.TypedObject
		derived := Must(runtime.NewDefaultScheme(&rt, &runtime.UnstructuredTypedObject{}, false, nil, scheme))
		Expect(Must(derived.Decode(t1data, nil))).To(Equal(t1))
	})

	It("extends derived scheme", func() {
		var rt runtime.TypedObject
		derived := Must(runtime.NewDefaultScheme(&rt, &runtime.UnstructuredTypedObject{}, false, nil, scheme))
		MustBeSuccessful(derived.RegisterByDecoder("t2", T2Decoder))
		Expect(Must(derived.Decode(t2data, nil))).To(Equal(t2))

		Expect(scheme.KnownTypeNames()).To(Equal([]string{"t1"}))
		Expect(derived.KnownTypeNames()).To(Equal([]string{"t1", "t2"}))

		Expect(utils.StringMapKeys(scheme.KnownTypes())).To(Equal([]string{"t1"}))
		Expect(utils.StringMapKeys(derived.KnownTypes())).To(Equal([]string{"t1", "t2"}))
	})
})
