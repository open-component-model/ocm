//go:build windows
// +build windows

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
		var val string
		PathVarPF(flags, &val, "path", "p", "", "help message")
		flags.Parse([]string{"-p", `E:\t\bugrepo\postgresql-14.0.5.tgz`})
		Expect(val).To(Equal("E:/t/bugrepo/postgresql-14.0.5.tgz"))
	})

	It("parse default path", func() {
		var val string
		PathVarP(flags, &val, "path", "p", `E:\t\bugrepo\postgresql-14.0.5.tgz`, "help message")
		Expect(val).To(Equal("E:/t/bugrepo/postgresql-14.0.5.tgz"))
	})

})
