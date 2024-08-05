//go:build windows

package flag_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/utils/cobrautils/flag"

	"github.com/spf13/pflag"
)

var _ = Describe("path flags", func() {
	var flags *pflag.FlagSet

	BeforeEach(func() {
		flags = pflag.NewFlagSet("test", pflag.ContinueOnError)
	})

	It("parse windows path", func() {
		var val []string
		PathArrayVarPF(flags, &val, "path", "p", nil, "help message")
		flags.Parse([]string{"-p", `C:\foo\bar;E:\other\path`})
		Expect(val).To(Equal([]string{"C:/foo/bar", "E:/other/path"}))
	})

	It("parse default path", func() {
		var val []string
		PathArrayVarPF(flags, &val, "path", "p", []string{`C:\foo\bar`, `E:\other\path`}, "help message")
		Expect(val).To(Equal([]string{"C:/foo/bar", "E:/other/path"}))
	})
})
