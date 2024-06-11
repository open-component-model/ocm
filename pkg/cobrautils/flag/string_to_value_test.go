package flag

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/goutils/testutils"
	"github.com/spf13/pflag"
)

var _ = Describe("value map", func() {
	var flags *pflag.FlagSet

	BeforeEach(func() {
		flags = pflag.NewFlagSet("test", pflag.ContinueOnError)
	})

	It("handles simple map content", func() {
		var flag map[string]interface{}
		StringToValueVarP(flags, &flag, "flag", "", nil, "test flag")

		value := `a=b`

		Expect(flags.Parse([]string{"--flag", value})).To(Succeed())
		Expect(flag).To(Equal(map[string]interface{}{"a": "b"}))
	})

	It("handles generic yaml content", func() {
		var flag map[string]interface{}
		StringToValueVarP(flags, &flag, "flag", "", nil, "test flag")

		value := `a={"a":"va"}`

		Expect(flags.Parse([]string{"--flag", value})).To(Succeed())
		Expect(flag).To(Equal(map[string]interface{}{"a": map[string]interface{}{"a": "va"}}))
	})

	It("shows default", func() {
		var flag map[string]interface{}
		StringToValueVarP(flags, &flag, "flag", "", map[string]interface{}{"x": map[string]interface{}{"a": "b"}}, "test flag")

		Expect(flags.FlagUsages()).To(testutils.StringEqualTrimmedWithContext(`--flag <name>=<YAML>   test flag (default [x={"a":"b"}])`))
	})

	It("handles replaces default content", func() {
		var flag map[string]interface{}
		StringToValueVarP(flags, &flag, "flag", "", map[string]interface{}{"x": "y"}, "test flag")

		value := `a=b`

		Expect(flags.Parse([]string{"--flag", value})).To(Succeed())
		Expect(flag).To(Equal(map[string]interface{}{"a": "b"}))
	})

	It("rejects invalid value", func() {
		var flag map[string]interface{}
		StringToValueVarP(flags, &flag, "flag", "", nil, "test flag")

		value := `a={"a":"va"`

		err := flags.Parse([]string{"--flag", value})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("invalid argument \"a={\\\"a\\\":\\\"va\\\"\" for \"--flag\" flag: failed to parse YAML: \"{\\\"a\\\":\\\"va\\\"\": error converting YAML to JSON: yaml: line 1: did not find expected ',' or '}'"))
	})

	It("rejects invalid assignment", func() {
		var flag map[string]interface{}
		StringToValueVarP(flags, &flag, "flag", "", nil, "test flag")

		value := `a`

		err := flags.Parse([]string{"--flag", value})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("invalid argument \"a\" for \"--flag\" flag: expected <name>=<value>"))
	})
})
