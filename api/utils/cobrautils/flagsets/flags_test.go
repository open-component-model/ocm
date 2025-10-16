package flagsets_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/pflag"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
)

var _ = Describe("type set", func() {
	var set flagsets.ConfigOptionTypeSet
	var flags *pflag.FlagSet

	BeforeEach(func() {
		set = flagsets.NewConfigOptionTypeSet("first")
		flags = pflag.NewFlagSet("flags", pflag.ContinueOnError)
	})

	It("handles option group in args", func() {
		set1 := flagsets.NewConfigOptionTypeSet("first")
		set1.AddOptionType(flagsets.NewStringOptionType("string", "a test string"))
		set1.AddOptionType(flagsets.NewStringOptionType("other", "another test string"))

		set2 := flagsets.NewConfigOptionTypeSet("second")
		set2.AddOptionType(flagsets.NewStringOptionType("string", "a test string"))
		set2.AddOptionType(flagsets.NewStringOptionType("third", "a third string"))

		Expect(set.AddTypeSet(set1)).To(Succeed())
		Expect(set.AddTypeSet(set2)).To(Succeed())

		opts := set.CreateOptions()

		opts.AddFlags(flags)

		Expect(flags.Parse([]string{"--string=string", "--other=other"})).To(Succeed())

		Expect(opts.Check(set.GetTypeSet("first"), "")).To(Succeed())
	})

	It("fails for mixed option group in args", func() {
		set1 := flagsets.NewConfigOptionTypeSet("first")
		set1.AddOptionType(flagsets.NewStringOptionType("string", "a test string"))
		set1.AddOptionType(flagsets.NewStringOptionType("other", "another test string"))

		set2 := flagsets.NewConfigOptionTypeSet("second")
		set2.AddOptionType(flagsets.NewStringOptionType("string", "a test string"))
		set2.AddOptionType(flagsets.NewStringOptionType("third", "a third string"))

		Expect(set.AddTypeSet(set1)).To(Succeed())
		Expect(set.AddTypeSet(set2)).To(Succeed())

		opts := set.CreateOptions()
		opts.AddFlags(flags)

		Expect(flags.Parse([]string{"--string=string", "--other=other", "--third=third"})).To(Succeed())

		err := opts.Check(set.GetTypeSet("first"), "")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("option \"third\" given, but not valid for option set \"first\""))
	})

	It("provides config value", func() {
		set.AddOptionType(flagsets.NewStringOptionType("string", "a test string"))
		set.AddOptionType(flagsets.NewStringOptionType("other", "another test string"))

		opts := set.CreateOptions()
		opts.AddFlags(flags)

		Expect(flags.Parse([]string{"--string=string", "--other=other string"})).To(Succeed())

		v, ok := opts.GetValue("string")
		Expect(ok).To(BeTrue())
		Expect(v.(string)).To(Equal("string"))

		v, ok = opts.GetValue("other")
		Expect(ok).To(BeTrue())
		Expect(v.(string)).To(Equal("other string"))
	})
})
