package flag

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/spf13/pflag"
)

var _ = Describe("yaml flags", func() {
	var flags *pflag.FlagSet

	BeforeEach(func() {
		flags = pflag.NewFlagSet("test", pflag.ContinueOnError)
	})

	It("handles generic yaml content", func() {
		var flag LabelledString
		LabelledStringVarP(flags, &flag, "flag", "", LabelledString{}, "test flag")

		value := `a=b`

		Expect(flags.Parse([]string{"--flag", value})).To(Succeed())
		Expect(flag).To(Equal(LabelledString{"a", "b"}))
	})

	It("rejects invalid assignment", func() {
		var flag LabelledString
		LabelledStringVarP(flags, &flag, "flag", "", LabelledString{}, "test flag")

		value := `a`

		err := flags.Parse([]string{"--flag", value})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("invalid argument \"a\" for \"--flag\" flag: expected <name>=<value>"))
	})
})
