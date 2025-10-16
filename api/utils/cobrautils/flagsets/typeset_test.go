package flagsets_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/utils/cobrautils/flagsets"
)

var _ = Describe("type set", func() {
	var set flagsets.ConfigOptionTypeSet

	BeforeEach(func() {
		set = flagsets.NewConfigOptionTypeSet("set")
	})

	It("composes type set", func() {
		set.AddOptionType(flagsets.NewStringOptionType("string", "a test string"))
		set.AddOptionType(flagsets.NewStringOptionType("other", "another test string"))

		Expect(set.GetOptionType("string")).To(Equal(flagsets.NewStringOptionType("string", "a test string")))
		Expect(set.GetOptionType("other")).To(Equal(flagsets.NewStringOptionType("other", "another test string")))
	})

	It("aligns two sets", func() {
		set1 := flagsets.NewConfigOptionTypeSet("first")
		set1.AddOptionType(flagsets.NewStringOptionType("string", "a test string"))
		set1.AddOptionType(flagsets.NewStringOptionType("other", "another test string"))

		set2 := flagsets.NewConfigOptionTypeSet("second")
		set2.AddOptionType(flagsets.NewStringOptionType("string", "a test string"))
		set2.AddOptionType(flagsets.NewStringOptionType("third", "a third string"))

		Expect(set.AddTypeSet(set1)).To(Succeed())
		Expect(set.AddTypeSet(set2)).To(Succeed())

		Expect(set.GetOptionType("string")).To(Equal(flagsets.NewStringOptionType("string", "a test string")))
		Expect(set.GetOptionType("other")).To(Equal(flagsets.NewStringOptionType("other", "another test string")))
		Expect(set.GetOptionType("third")).To(Equal(flagsets.NewStringOptionType("third", "a third string")))

		Expect(set.GetOptionType("string")).To(BeIdenticalTo(set1.GetOptionType("string")))
		Expect(set.GetOptionType("string")).NotTo(BeIdenticalTo(set2.GetOptionType("string")))
		Expect(set.GetOptionType("string")).To(Equal(set2.GetOptionType("string")))
		Expect(set.GetOptionType("other")).To(BeIdenticalTo(set1.GetOptionType("other")))
		Expect(set.GetOptionType("third")).To(BeIdenticalTo(set2.GetOptionType("third")))
	})
})
