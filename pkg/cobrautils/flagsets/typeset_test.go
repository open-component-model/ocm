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

package flagsets_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
)

var _ = Describe("type set", func() {
	var set flagsets.ConfigOptionTypeSet

	BeforeEach(func() {
		set = flagsets.NewConfigOptionSet("set")
	})

	It("composes type set", func() {
		set.AddOptionType(flagsets.NewStringOptionType("string", "a test string"))
		set.AddOptionType(flagsets.NewStringOptionType("other", "another test string"))

		Expect(set.GetOptionType("string")).To(Equal(flagsets.NewStringOptionType("string", "a test string")))
		Expect(set.GetOptionType("other")).To(Equal(flagsets.NewStringOptionType("other", "another test string")))
	})

	It("aligns two sets", func() {
		set1 := flagsets.NewConfigOptionSet("first")
		set1.AddOptionType(flagsets.NewStringOptionType("string", "a test string"))
		set1.AddOptionType(flagsets.NewStringOptionType("other", "another test string"))

		set2 := flagsets.NewConfigOptionSet("second")
		set2.AddOptionType(flagsets.NewStringOptionType("string", "a test string"))
		set2.AddOptionType(flagsets.NewStringOptionType("third", "a third string"))

		Expect(set.AddTypeSet(set1)).To(Succeed())
		Expect(set.AddTypeSet(set2)).To(Succeed())

		Expect(set.GetOptionType("string")).To(Equal(flagsets.NewStringOptionType("string", "a test string")))
		Expect(set.GetOptionType("other")).To(Equal(flagsets.NewStringOptionType("other", "another test string")))
		Expect(set.GetOptionType("third")).To(Equal(flagsets.NewStringOptionType("third", "a third string")))

		Expect(set.GetOptionType("string")).To(BeIdenticalTo(set1.GetOptionType("string")))
		Expect(set.GetOptionType("string")).To(BeIdenticalTo(set2.GetOptionType("string")))
		Expect(set.GetOptionType("other")).To(BeIdenticalTo(set1.GetOptionType("other")))
		Expect(set.GetOptionType("third")).To(BeIdenticalTo(set2.GetOptionType("third")))
	})
})
