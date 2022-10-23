// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

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
