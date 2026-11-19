package groups_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/spf13/pflag"

	"ocm.software/ocm/api/utils/cobrautils/groups"
)

// groupUsage returns the rendered usages for the group with the given title.
func groupUsage(r []groups.UsageGroup, title string) (string, bool) {
	for _, g := range r {
		if g.Title == title {
			return g.Usages, true
		}
	}
	return "", false
}

var _ = Describe("FlagUsages", func() {
	var fs *pflag.FlagSet

	BeforeEach(func() {
		fs = pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.StringP("date", "d", "", "the date to ask for (MM/DD)")
		fs.BoolP("help", "h", false, "help for check")
	})

	It("renders bool flags without the pflag v1.1.0 value suffix", func() {
		out := groups.FlagUsagesWrapped(fs, 0)
		Expect(out).To(ContainSubstring("--help"))
		Expect(out).NotTo(ContainSubstring("[=true|false]"))
	})

	It("keeps bool parse semantics after rendering", func() {
		groups.FlagUsagesWrapped(fs, 0)

		Expect(fs.Parse([]string{"-h"})).To(Succeed())
		Expect(fs.Lookup("help").Value.String()).To(Equal("true"))

		fs2 := pflag.NewFlagSet("test2", pflag.ContinueOnError)
		fs2.Bool("help", false, "help")
		groups.FlagUsagesWrapped(fs2, 0)
		Expect(fs2.Parse([]string{"--help=false"})).To(Succeed())
		Expect(fs2.Lookup("help").Value.String()).To(Equal("false"))
	})

	It("attributes multi-line descriptions to all of a flag's groups", func() {
		fs.Bool("verbose", false, "first marker\n\nsecond marker\nsee --date for detail")
		Expect(fs.SetAnnotation("verbose", groups.FlagGroupAnnotation, []string{"Output", "Diagnostics"})).To(Succeed())

		r := groups.GroupedFlagUsagesWrapped(fs, 0)

		def, ok := groupUsage(r, "")
		Expect(ok).To(BeTrue())
		Expect(def).To(ContainSubstring("--date"))
		Expect(def).To(ContainSubstring("--help"))
		// the verbose continuation must not leak into the default group
		Expect(def).NotTo(ContainSubstring("--verbose"))
		Expect(def).NotTo(ContainSubstring("first marker"))
		Expect(def).NotTo(ContainSubstring("second marker"))

		for _, title := range []string{"Output", "Diagnostics"} {
			usage, ok := groupUsage(r, title)
			Expect(ok).To(BeTrue(), "group %q present", title)
			Expect(usage).To(ContainSubstring("--verbose"))
			// every supplied description line, including the blank one, survives
			Expect(usage).To(ContainSubstring("first marker"))
			Expect(usage).To(ContainSubstring("second marker"))
			Expect(usage).To(ContainSubstring("see --date for detail"))
			// the blank separator line between the two markers survives
			Expect(usage).To(MatchRegexp(`first marker\n\s*\n\s*second marker`))
			Expect(usage).NotTo(ContainSubstring("[=true|false]"))
		}
	})

	It("keeps wrapped continuation lines with their flag's group", func() {
		fs.Bool("verbose", false,
			"alpha bravo charlie delta echo foxtrot golf hotel india juliet kilo lima --date november oscar papa quebec romeo")
		Expect(fs.SetAnnotation("verbose", groups.FlagGroupAnnotation, []string{"Output"})).To(Succeed())

		r := groups.GroupedFlagUsagesWrapped(fs, 60)

		usage, ok := groupUsage(r, "Output")
		Expect(ok).To(BeTrue())
		// the long usage wraps onto more than one line
		Expect(len(splitNonEmpty(usage))).To(BeNumerically(">", 1))
		for _, tok := range []string{"alpha", "lima", "--date", "november", "romeo"} {
			Expect(usage).To(ContainSubstring(tok))
		}

		// the wrapped "--date" token must not reassign ownership to another group
		def, ok := groupUsage(r, "")
		Expect(ok).To(BeTrue())
		Expect(def).NotTo(ContainSubstring("november"))
		Expect(def).NotTo(ContainSubstring("romeo"))
	})

	It("keeps an annotated non-default NoOptDefVal flag in its group", func() {
		fs.Bool("flag", true, "x")
		fs.Lookup("flag").NoOptDefVal = "false"
		Expect(fs.SetAnnotation("flag", groups.FlagGroupAnnotation, []string{"Output"})).To(Succeed())

		usage, ok := groupUsage(groups.GroupedFlagUsagesWrapped(fs, 0), "Output")
		Expect(ok).To(BeTrue())
		Expect(usage).To(ContainSubstring("--flag[=false]"))
		Expect(usage).To(ContainSubstring("(default true)"))
		Expect(usage).NotTo(ContainSubstring("[=true|false]"))
	})

	It("leaves the original flag set unmodified", func() {
		groups.FlagUsagesWrapped(fs, 0)

		// the original value still renders in plain pflag v1.1.0 format
		Expect(fs.FlagUsages()).To(ContainSubstring("--help[=true|false]"))
	})

	It("yields nothing for an empty flag set", func() {
		empty := pflag.NewFlagSet("empty", pflag.ContinueOnError)
		Expect(groups.DetermineGroups(empty, 0)).To(BeEmpty())
		Expect(groups.GroupedFlagUsagesWrapped(empty, 0)).To(BeEmpty())
		Expect(groups.FlagUsagesWrapped(empty, 0)).To(Equal(""))
	})
})

func splitNonEmpty(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if l != "" {
			out = append(out, l)
		}
	}
	return out
}
