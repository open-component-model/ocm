// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package runtime_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/runtime"
)

type TestSpecType = runtime.VersionedTypedObjectType[TestSpecRealm]

var _ = Describe("multi version type", func() {
	scheme := runtime.MustNewDefaultTypeScheme[TestSpecRealm, TestSpecType](nil, false, nil)

	versions := runtime.NewTypeVersionScheme[TestSpecRealm, TestSpecType](Type1, scheme)

	versions.Register(runtime.NewVersionedTypedObjectTypeByConverter[TestSpecRealm, *TestSpec1, *Spec1V1](Type1, &converterSpec1V1{}))
	versions.Register(runtime.NewVersionedTypedObjectTypeByConverter[TestSpecRealm, *TestSpec1, *Spec1V1](Type1V1, &converterSpec1V1{}))
	versions.Register(runtime.NewVersionedTypedObjectTypeByConverter[TestSpecRealm, *TestSpec1, *Spec1V2](Type1V2, &converterSpec1V2{}))

	multi := Must(runtime.NewMultiFormatVersionedType[TestSpecRealm, TestSpecType](Type1, versions))

	It("decodes plain version wth v2", func() {
		s := `
type: ` + Type1 + `
field: sally
`
		spec := Must(multi.Decode([]byte(s), nil))

		Expect(spec.GetType()).To(Equal(Type1))
		Expect(spec.(*TestSpec1).Field).To(Equal("sally"))
	})

	It("decodes plain version with v1", func() {
		s := `
type: ` + Type1 + `
oldField: sally
`
		spec := Must(multi.Decode([]byte(s), nil))

		Expect(spec.GetType()).To(Equal(Type1))
		Expect(spec.(*TestSpec1).Field).To(Equal("sally"))
	})
})
