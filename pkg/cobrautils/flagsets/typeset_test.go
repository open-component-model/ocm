// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package flagsets_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
)

var _ = Describe("type set", func() {
	var set flagsets.ConfigOptionTypeSet

	BeforeEach(func() {
		set = flagsets.NewConfigOptionSet("set")
	})

	It("composes type set", func() {
		set.AddOptionType(flagsets.NewStringOptionType("string", "a test string"))
		set.AddOptionType(flagsets.NewStringOptionType("other", "another test string"))

		Expect(set.GetOptionType("string")).To(Equal(flagsets.NewStringOptionType("string", "a test string")))
		Expect(set.GetOptionType("other")).To(Equal(flagsets.NewStringOptionType("other", "another test string")))
	})

	It("aligns two sets", func() {
		set1 := flagsets.NewConfigOptionSet("first")
		set1.AddOptionType(flagsets.NewStringOptionType("string", "a test string"))
		set1.AddOptionType(flagsets.NewStringOptionType("other", "another test string"))

		set2 := flagsets.NewConfigOptionSet("second")
		set2.AddOptionType(flagsets.NewStringOptionType("string", "a test string"))
		set2.AddOptionType(flagsets.NewStringOptionType("third", "a third string"))

		Expect(set.AddTypeSet(set1)).To(Succeed())
		Expect(set.AddTypeSet(set2)).To(Succeed())

		Expect(set.GetOptionType("string")).To(Equal(flagsets.NewStringOptionType("string", "a test string")))
		Expect(set.GetOptionType("other")).To(Equal(flagsets.NewStringOptionType("other", "another test string")))
		Expect(set.GetOptionType("third")).To(Equal(flagsets.NewStringOptionType("third", "a third string")))

		Expect(set.GetOptionType("string")).To(BeIdenticalTo(set1.GetOptionType("string")))
		Expect(set.GetOptionType("string")).To(BeIdenticalTo(set2.GetOptionType("string")))
		Expect(set.GetOptionType("other")).To(BeIdenticalTo(set1.GetOptionType("other")))
		Expect(set.GetOptionType("third")).To(BeIdenticalTo(set2.GetOptionType("third")))
	})
})
