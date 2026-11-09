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
		root.PersistentFlags().Bool("verbose", false, "verbose output")
		child = &cobra.Command{Use: "child", Short: "child command", Run: func(*cobra.Command, []string) {}}
		child.Flags().Bool("local", false, "local flag")
		root.AddCommand(child)
		cobrautils.TweakCommand(root, nil)
	})

	It("renders inherited bool flags without value suffix (help)", func() {
		buf := &bytes.Buffer{}
		root.SetOut(buf)
		root.SetArgs([]string{"child", "--help"})
		Expect(root.Execute()).To(Succeed())
		Expect(buf.String()).To(ContainSubstring("Global Options:"))
		Expect(buf.String()).To(MatchRegexp(`(?m)^\s+--verbose\s+verbose output$`))
		Expect(buf.String()).NotTo(ContainSubstring("[=true|false]"))
	})

	It("renders inherited bool flags without value suffix (usage)", func() {
		u := child.UsageString()
		Expect(u).To(ContainSubstring("Global Options:"))
		Expect(u).To(MatchRegexp(`(?m)^\s+--verbose\s+verbose output$`))
		Expect(u).NotTo(ContainSubstring("[=true|false]"))
	})
})
