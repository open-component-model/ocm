// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package defaultmerge

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/runtime"
)

var _ = Describe("list merge", func() {
	handler := New()

	var e1, e2 Value
	var a, b runtime.RawValue

	BeforeEach(func() {
		e1 = "v1"
		e2 = "v2"

		MustBeSuccessful(a.SetValue(e1))
		MustBeSuccessful(b.SetValue(e1))
	})

	It("merges no change", func() {
		MustBeSuccessful(handler.Merge(nil, a, &b, nil))
		Expect(b).To(Equal(a))
	})

	It("modified keeps local", func() {
		MustBeSuccessful(a.SetValue(e2))
		MustBeSuccessful(handler.Merge(nil, a, &b, NewConfig(MODE_LOCAL)))

		Expect(b).To(DeepEqual(a))
	})

	It("modified accept inbound", func() {
		MustBeSuccessful(b.SetValue(e2))
		r := b.Copy()
		MustBeSuccessful(handler.Merge(nil, a, &b, nil))
		Expect(b).To(Equal(r))
	})

	It("fails for none mode", func() {
		MustBeSuccessful(b.SetValue(e2))
		MustFailWithMessage(ErrorFrom(handler.Merge(nil, a, &b, NewConfig(MODE_NONE))), "[default]: target value changed")
	})
})
