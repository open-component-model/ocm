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

	It("handles string yaml content", func() {
		var flag LabelledValue
		LabelledValueVarP(flags, &flag, "flag", "", LabelledValue{}, "test flag")

		value := `a=b`

		Expect(flags.Parse([]string{"--flag", value})).To(Succeed())
		Expect(flag).To(Equal(LabelledValue{"a", "b"}))
	})

	It("handles generic yaml content", func() {
		var flag LabelledValue
		LabelledValueVarP(flags, &flag, "flag", "", LabelledValue{}, "test flag")

		value := `a={"a":"va"}`

		Expect(flags.Parse([]string{"--flag", value})).To(Succeed())
		Expect(flag).To(Equal(LabelledValue{"a", map[string]interface{}{"a": "va"}}))
	})

	It("rejects invalid value", func() {
		var flag LabelledValue
		LabelledValueVarP(flags, &flag, "flag", "", LabelledValue{}, "test flag")

		value := `a={"a":"va"`

		err := flags.Parse([]string{"--flag", value})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("invalid argument \"a={\\\"a\\\":\\\"va\\\"\" for \"--flag\" flag: error converting YAML to JSON: yaml: line 1: did not find expected ',' or '}'"))
	})

	It("rejects invalid assignment", func() {
		var flag LabelledValue
		LabelledValueVarP(flags, &flag, "flag", "", LabelledValue{}, "test flag")

		value := `a`

		err := flags.Parse([]string{"--flag", value})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("invalid argument \"a\" for \"--flag\" flag: expected <name>=<value>"))
	})
})
