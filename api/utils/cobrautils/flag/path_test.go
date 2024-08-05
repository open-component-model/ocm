//go:build !windows

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

	It("parse linux path", func() {
		var val string
		PathVarP(flags, &val, "path", "p", "", "help message")
		flags.Parse([]string{"-p", `/t/bugrepo/postgresql-14.0.5.tgz`})
		Expect(val).To(Equal("/t/bugrepo/postgresql-14.0.5.tgz"))
	})

	It("parse default path", func() {
		var val string
		PathVarP(flags, &val, "path", "p", `/t/bugrepo/postgresql-14.0.5.tgz`, "help message")
		Expect(val).To(Equal("/t/bugrepo/postgresql-14.0.5.tgz"))
	})
})
