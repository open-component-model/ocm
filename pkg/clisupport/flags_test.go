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

package clisupport

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/spf13/pflag"
)

var _ = Describe("type set", func() {
	var set ConfigOptionTypeSet
	var flags *pflag.FlagSet

	BeforeEach(func() {
		set = NewConfigOptionSet("first")
		flags = pflag.NewFlagSet("flags", pflag.ContinueOnError)
	})

	It("handles option group in args", func() {
		set1 := NewConfigOptionSet("first")
		set1.AddOptionType(NewStringOptionType("string", "a test string"))
		set1.AddOptionType(NewStringOptionType("other", "another test string"))

		set2 := NewConfigOptionSet("second")
		set2.AddOptionType(NewStringOptionType("string", "a test string"))
		set2.AddOptionType(NewStringOptionType("third", "a third string"))

		Expect(set.AddTypeSet(set1)).To(Succeed())
		Expect(set.AddTypeSet(set2)).To(Succeed())

		opts := set.CreateOptions()

		opts.AddFlags(flags)

		Expect(flags.Parse([]string{"--string=string", "--other=other"})).To(Succeed())

		Expect(opts.Check(set.GetTypeSet("first"), "")).To(Succeed())
	})

	It("fails for mixed option group in args", func() {
		set1 := NewConfigOptionSet("first")
		set1.AddOptionType(NewStringOptionType("string", "a test string"))
		set1.AddOptionType(NewStringOptionType("other", "another test string"))

		set2 := NewConfigOptionSet("second")
		set2.AddOptionType(NewStringOptionType("string", "a test string"))
		set2.AddOptionType(NewStringOptionType("third", "a third string"))

		Expect(set.AddTypeSet(set1)).To(Succeed())
		Expect(set.AddTypeSet(set2)).To(Succeed())

		opts := set.CreateOptions()
		opts.AddFlags(flags)

		Expect(flags.Parse([]string{"--string=string", "--other=other", "--third=third"})).To(Succeed())

		err := opts.Check(set.GetTypeSet("first"), "")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("option \"third\" given, but not valid for option set \"first\""))
	})

	It("provides config value", func() {
		set.AddOptionType(NewStringOptionType("string", "a test string"))
		set.AddOptionType(NewStringOptionType("other", "another test string"))

		opts := set.CreateOptions()
		opts.AddFlags(flags)

		Expect(flags.Parse([]string{"--string=string", "--other=other string"})).To(Succeed())

		v, ok := opts.GetValue("string")
		Expect(ok).To(BeTrue())
		Expect(v.(string)).To(Equal("string"))

		v, ok = opts.GetValue("other")
		Expect(ok).To(BeTrue())
		Expect(v.(string)).To(Equal("other string"))
	})
})
