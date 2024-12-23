package flag

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/goutils/testutils"
	"github.com/spf13/pflag"
)

var _ = Describe("string map slice", func() {
	var flags *pflag.FlagSet

	BeforeEach(func() {
		flags = pflag.NewFlagSet("test", pflag.ContinueOnError)
	})

	It("handles simple identity", func() {
		var flag []map[string]string
		StringMapSliceVarP(flags, &flag, "type", "flag", "", nil, "test flag")

		value := `type=alice`

		Expect(flags.Parse([]string{"--flag", value})).To(Succeed())
		Expect(flag).To(Equal([]map[string]string{{"type": "alice"}}))
	})

	It("handles simple path", func() {
		var flag []map[string]string
		StringMapSliceVarP(flags, &flag, "type", "flag", "", nil, "test flag")

		value1 := `type=alice`
		value2 := `husband=bob`

		Expect(flags.Parse([]string{"--flag", value1, "--flag", value2})).To(Succeed())
		Expect(flag).To(Equal([]map[string]string{{"type": "alice", "husband": "bob"}}))
	})

	It("handles multi path", func() {
		var flag []map[string]string
		StringMapSliceVarP(flags, &flag, "type", "flag", "", nil, "test flag")

		value1 := `type=alice`
		value2 := `husband=bob`
		value3 := "type=bob"
		value4 := "wife=alice"
		value5 := "type=other"
		Expect(flags.Parse([]string{"--flag", value1, "--flag", value2, "--flag", value3, "--flag", value4, "--flag", value5})).To(Succeed())
		Expect(flag).To(Equal([]map[string]string{{"type": "alice", "husband": "bob"}, {"type": "bob", "wife": "alice"}, {"type": "other"}}))
	})

	It("shows default", func() {
		var flag []map[string]string
		StringMapSliceVarP(flags, &flag, "type", "flag", "", []map[string]string{{"type": "alice"}}, "test flag")

		Expect(flags.FlagUsages()).To(testutils.StringEqualTrimmedWithContext(`--flag {<name[type]>=<value>}   test flag (default [{"type":"alice"}])`))
	})

	It("handles replaces default content", func() {
		var flag []map[string]string
		StringMapSliceVarP(flags, &flag, "type", "flag", "", []map[string]string{{"name": "other"}}, "test flag")

		value1 := `type=alice`
		value2 := `husband=bob`

		Expect(flags.Parse([]string{"--flag", value1, "--flag", value2})).To(Succeed())
		Expect(flag).To(Equal([]map[string]string{{"type": "alice", "husband": "bob"}}))
	})

	It("rejects invalid value", func() {
		var flag []map[string]string
		StringMapSliceVarP(flags, &flag, "type", "flag", "", nil, "test flag")

		value := `a=b`

		err := flags.Parse([]string{"--flag", value})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("invalid argument \"a=b\" for \"--flag\" flag: first attribute must be the \"type\" attribute"))
	})

	It("simplified main", func() {
		var flag []map[string]string
		StringMapSliceVarP(flags, &flag, "type", "flag", "", nil, "test flag")

		value := `alice`

		err := flags.Parse([]string{"--flag", value})
		Expect(err).To(Succeed())
		Expect(flag).To(Equal([]map[string]string{{"type": "alice"}}))
	})

	/*
		It("rejects invalid assignment", func() {
			var flag []map[string]string
			StringMapSliceVarP(flags, &flag, "type", "flag", "", nil, "test flag")

			value := `a`

			err := flags.Parse([]string{"--flag", value})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid argument \"a\" for \"--flag\" flag: expected <name>=<value>"))
		})
	*/
})
