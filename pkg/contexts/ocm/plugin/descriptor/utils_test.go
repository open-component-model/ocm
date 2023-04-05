// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package descriptor

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("utils", func() {

	It("merges lists", func() {
		l1 := List[StringName]{"a", "b"}
		l2 := List[StringName]{"b", "c"}
		l3 := List[StringName]{"a", "d"}

		m2 := l1.MergeWith(l2)
		Expect(m2).To(Equal(List[StringName]{"a", "b", "c"}))
		Expect(l1.MergeWith(l3)).To(Equal(List[StringName]{"a", "b", "d"}))
		Expect(m2).To(Equal(List[StringName]{"a", "b", "c"}))
	})
})
