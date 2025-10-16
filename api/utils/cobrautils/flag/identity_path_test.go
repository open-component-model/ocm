package flag

import (
	"github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/pflag"
)

var _ = Describe("identity path", func() {
	var flags *pflag.FlagSet

	BeforeEach(func() {
		flags = pflag.NewFlagSet("test", pflag.ContinueOnError)
	})

	It("handles simple identity", func() {
		var flag []map[string]string
		IdentityPathVarP(flags, &flag, "flag", "", nil, "test flag")

		value := `name=alice`

		Expect(flags.Parse([]string{"--flag", value})).To(Succeed())
		Expect(flag).To(Equal([]map[string]string{{"name": "alice"}}))
	})

	It("handles simple path", func() {
		var flag []map[string]string
		IdentityPathVarP(flags, &flag, "flag", "", nil, "test flag")

		value1 := `name=alice`
		value2 := `husband=bob`

		Expect(flags.Parse([]string{"--flag", value1, "--flag", value2})).To(Succeed())
		Expect(flag).To(Equal([]map[string]string{{"name": "alice", "husband": "bob"}}))
	})

	It("handles multi path", func() {
		var flag []map[string]string
		IdentityPathVarP(flags, &flag, "flag", "", nil, "test flag")

		value1 := `name=alice`
		value2 := `husband=bob`
		value3 := "name=bob"
		value4 := "wife=alice"
		value5 := "name=other"
		Expect(flags.Parse([]string{"--flag", value1, "--flag", value2, "--flag", value3, "--flag", value4, "--flag", value5})).To(Succeed())
		Expect(flag).To(Equal([]map[string]string{{"name": "alice", "husband": "bob"}, {"name": "bob", "wife": "alice"}, {"name": "other"}}))
	})

	It("shows default", func() {
		var flag []map[string]string
		IdentityPathVarP(flags, &flag, "flag", "", []map[string]string{{"name": "alice"}}, "test flag")

		Expect(flags.FlagUsages()).To(testutils.StringEqualTrimmedWithContext(`--flag {<name>=<value>}   test flag (default [{"name":"alice"}])`))
	})

	It("handles replaces default content", func() {
		var flag []map[string]string
		IdentityPathVarP(flags, &flag, "flag", "", []map[string]string{{"name": "other"}}, "test flag")

		value1 := `name=alice`
		value2 := `husband=bob`

		Expect(flags.Parse([]string{"--flag", value1, "--flag", value2})).To(Succeed())
		Expect(flag).To(Equal([]map[string]string{{"name": "alice", "husband": "bob"}}))
	})

	It("rejects invalid value", func() {
		var flag []map[string]string
		IdentityPathVarP(flags, &flag, "flag", "", nil, "test flag")

		value := `a=b`

		err := flags.Parse([]string{"--flag", value})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("invalid argument \"a=b\" for \"--flag\" flag: first identity attribute must be the name attribute"))
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
