package groups_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/goutils/testutils"
	"github.com/spf13/pflag"

	"ocm.software/ocm/api/utils/cobrautils/groups"
)

var _ = Describe("FlagUsages", func() {
	var fs *pflag.FlagSet

	BeforeEach(func() {
		fs = pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.StringP("date", "d", "", "the date to ask for (MM/DD)")
		fs.BoolP("help", "h", false, "help for check")
	})

	It("renders bool flags without value suffix", func() {
		Expect(groups.FlagUsagesWrapped(fs, 0)).To(testutils.StringEqualTrimmedWithContext(`
  -d, --date string   the date to ask for (MM/DD)
  -h, --help          help for check
`))
	})

	It("groups annotated bool flags", func() {
		fs.Bool("verbose", false, "verbose output")
		Expect(fs.SetAnnotation("verbose", groups.FlagGroupAnnotation, []string{"Output"})).To(Succeed())

		r := groups.GroupedFlagUsagesWrapped(fs, 0)
		Expect(r).To(HaveLen(2))

		// default group (empty title) has date and help
		Expect(r[0].Title).To(Equal(""))
		Expect(r[0].Usages).To(ContainSubstring("--date"))
		Expect(r[0].Usages).To(ContainSubstring("--help"))

		// Output group has verbose, no [=true|false]
		Expect(r[1].Title).To(Equal("Output"))
		Expect(r[1].Usages).To(ContainSubstring("--verbose"))
		Expect(r[1].Usages).NotTo(ContainSubstring("[=true|false]"))
	})

	It("renders bool flags with non-default no-opt value", func() {
		fs.Bool("flag", true, "x")
		fs.Lookup("flag").NoOptDefVal = "false"

		result := groups.FlagUsagesWrapped(fs, 0)
		Expect(result).To(ContainSubstring("--flag[=false]"))
	})

	It("does not modify the original flag set", func() {
		groups.FlagUsagesWrapped(fs, 0)
		Expect(fs.Parse([]string{"-h"})).To(Succeed())
		Expect(fs.Lookup("help").Value.String()).To(Equal("true"))
	})
})
