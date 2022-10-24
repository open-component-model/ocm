// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package flag

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/pkg/testutils"
)

var _ = Describe("string maps", func() {
	var flags *pflag.FlagSet

	BeforeEach(func() {
		flags = pflag.NewFlagSet("test", pflag.ContinueOnError)
	})

	It("handles map content", func() {
		var flag map[string]string
		StringMapVarP(flags, &flag, "flag", "", nil, "test flag")

		value := `a=b`

		Expect(flags.Parse([]string{"--flag", value})).To(Succeed())
		Expect(flag).To(Equal(map[string]string{"a": "b"}))
	})

	It("shows default", func() {
		var flag map[string]string
		StringMapVarP(flags, &flag, "flag", "", map[string]string{"x": "y"}, "test flag")

		Expect(flags.FlagUsages()).To(testutils.StringEqualTrimmedWithContext("--flag <name>=<value>   test flag (default [x=y])"))
	})

	It("handles replaces default content", func() {
		var flag map[string]string
		StringMapVarP(flags, &flag, "flag", "", map[string]string{"x": "y"}, "test flag")

		value := `a=b`

		Expect(flags.Parse([]string{"--flag", value})).To(Succeed())
		Expect(flag).To(Equal(map[string]string{"a": "b"}))
	})

	It("rejects invalid assignment", func() {
		var flag map[string]string
		StringMapVarP(flags, &flag, "flag", "", nil, "test flag")

		value := `a`

		err := flags.Parse([]string{"--flag", value})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("invalid argument \"a\" for \"--flag\" flag: expected <name>=<value>"))
	})
})
