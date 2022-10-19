// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/logging"

	ocmlog "github.com/open-component-model/ocm/pkg/logging"
	"github.com/open-component-model/ocm/pkg/testutils"
)

var _ = Describe("Printer", func() {

	var buf *bytes.Buffer
	var printer Printer

	BeforeEach(func() {
		buf = &bytes.Buffer{}
		printer = NewPrinter(buf)
	})

	It("print", func() {
		printer.Printf("test")
		Expect(buf.String()).To(Equal("test"))
	})
	It("prints line", func() {
		printer.Printf("test\n")
		Expect(buf.String()).To(Equal("test\n"))
	})
	It("prints lines", func() {
		printer.Printf("line\ntest\n")
		Expect(buf.String()).To(Equal("line\ntest\n"))
	})

	It("prints gap", func() {
		printer.Printf("line\n")
		p := printer.AddGap("  ")
		p.Printf("test\n")
		Expect(buf.String()).To(Equal("line\n  test\n"))
		p.Printf("next\n")
		Expect(buf.String()).To(Equal("line\n  test\n  next\n"))
		printer.Printf("back\n")
		Expect(buf.String()).To(Equal("line\n  test\n  next\nback\n"))
	})

	Context("loggging", func() {
		var buf *bytes.Buffer
		var logctx logging.Context
		var printer Printer

		BeforeEach(func() {
			logctx, buf = ocmlog.NewBufferedContext()
			printer = NewLoggingPrinter(logctx.Logger())
		})

		It("logs ", func() {
			for i := 1; i < 3; i++ {
				printer.Printf("line %d\n", i)
			}
			Expect(buf.String()).To(testutils.StringEqualTrimmedWithContext(`
V[3] line 1
V[3] line 2
`))
		})

		It("logs successive output", func() {
			for i := 1; i < 3; i++ {
				printer.Printf("test %d ", i)
			}
			printer.Printf("\n")
			Expect(buf.String()).To(testutils.StringEqualTrimmedWithContext(`
V[3] test 1 test 2
`))
		})
		It("logs multi line output", func() {
			printer.Printf("line 1\nline 2\n")
			Expect(buf.String()).To(testutils.StringEqualTrimmedWithContext(`
V[3] line 1
V[3] line 2
`))
		})

		It("flushes incomplete line", func() {
			printer.Printf("line 1\nline 2")
			Flush(printer)
			Expect(buf.String()).To(testutils.StringEqualTrimmedWithContext(`
V[3] line 1
V[3] line 2
`))
		})

	})
})
