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
		var flag interface{}
		YAMLVarP(flags, &flag, "flag", "", nil, "test flag")

		value := `{ "a": "va" }`

		Expect(flags.Parse([]string{"--flag", value})).To(Succeed())
		Expect(flag).To(Equal(map[string]interface{}{"a": "va"}))
	})

	It("handles typed content", func() {
		type T struct {
			A string
		}

		var flag T
		YAMLVarP(flags, &flag, "flag", "", T{}, "test flag")

		value := `{ "a": "va" }`

		Expect(flags.Parse([]string{"--flag", value})).To(Succeed())
		Expect(flag).To(Equal(T{"va"}))
	})

	It("handles typed content pointer", func() {
		type T struct {
			A string
		}

		var flag *T
		YAMLVarP(flags, &flag, "flag", "", nil, "test flag")

		value := `{ "a": "va" }`

		Expect(flags.Parse([]string{"--flag", value})).To(Succeed())
		Expect(flag).To(Equal(&T{"va"}))
	})
})
