//go:build !windows
// +build !windows

package flag_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/pflag"

	. "github.com/open-component-model/ocm/pkg/cobrautils/flag"
)

var _ = Describe("path flags", func() {
	var flags *pflag.FlagSet

	BeforeEach(func() {
		flags = pflag.NewFlagSet("test", pflag.ContinueOnError)
	})

	It("parse windows path", func() {
		var val []string
		PathArrayVarPF(flags, &val, "path", "p", nil, "help message")
		flags.Parse([]string{"-p", `/foo/bar:other/path`})
		Expect(val).To(Equal([]string{"/foo/bar", "other/path"}))
	})

	It("parse default path", func() {
		var val []string
		PathArrayVarPF(flags, &val, "path", "p", []string{`/foo/bar`, `other/path`}, "help message")
		Expect(val).To(Equal([]string{"/foo/bar", "other/path"}))
	})

})
