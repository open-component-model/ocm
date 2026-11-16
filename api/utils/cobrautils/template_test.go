package cobrautils_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/spf13/cobra"

	"ocm.software/ocm/api/utils/cobrautils"
)

var _ = Describe("help templates", func() {
	var root, child *cobra.Command

	BeforeEach(func() {
		root = &cobra.Command{Use: "root", Short: "root command"}
		root.PersistentFlags().Bool("verbose", false, "verbose output\ncontinuation marker\nsee --local for detail")
		child = &cobra.Command{Use: "child", Short: "child command", Run: func(*cobra.Command, []string) {}}
		child.Flags().Bool("local", false, "local flag")
		root.AddCommand(child)
		cobrautils.TweakCommand(root, nil)
	})

	assertInheritedUsage := func(out string) {
		Expect(out).To(ContainSubstring("Global Options:"))
		// local and inherited flags both present
		Expect(out).To(MatchRegexp(`(?m)^\s+--local\s+local flag$`))
		Expect(out).To(MatchRegexp(`(?m)^\s+--verbose\s+verbose output$`))
		// every supplied inherited description line survives
		Expect(out).To(ContainSubstring("continuation marker"))
		Expect(out).To(ContainSubstring("see --local for detail"))
		Expect(out).NotTo(ContainSubstring("[=true|false]"))
	}

	It("renders inherited multi-line bool flags without value suffix (help)", func() {
		buf := &bytes.Buffer{}
		root.SetOut(buf)
		root.SetArgs([]string{"child", "--help"})
		Expect(root.Execute()).To(Succeed())
		assertInheritedUsage(buf.String())
	})

	It("renders inherited multi-line bool flags without value suffix (usage)", func() {
		assertInheritedUsage(child.UsageString())
	})
})
