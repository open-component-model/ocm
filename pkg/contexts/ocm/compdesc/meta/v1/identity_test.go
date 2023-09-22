// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package v1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/equivalent/testhelper"

	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
)

var _ = Describe("identity", func() {
	Context("equivalence", func() {
		var a v1.Identity
		var b v1.Identity

		BeforeEach(func() {
			a = v1.NewIdentity("name", "extra", "extra")
			b = a.Copy()
		})

		It("detects equal", func() {
			CheckEquivalent(a.Equivalent(b))
		})

		It("detects different value", func() {
			b["name"] = "X"
			CheckNotLocalHashEqual(a.Equivalent(b))
			CheckNotLocalHashEqual(b.Equivalent(a))
		})

		It("detects additional attr", func() {
			b["X"] = "X"
			CheckNotLocalHashEqual(a.Equivalent(b))
			CheckNotLocalHashEqual(b.Equivalent(a))
		})

		It("detects replaced attr", func() {
			b["X"] = "extra"
			delete(b, "extra")
			Expect(len(a)).To(Equal(len(b)))
			CheckNotLocalHashEqual(a.Equivalent(b))
			CheckNotLocalHashEqual(b.Equivalent(a))
		})

	})
})
