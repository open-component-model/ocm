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

package options_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
)

type TestOption struct {
	Flag bool
}

func (t *TestOption) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&t.Flag, "flag", "f", false, "test flag")
}

var _ options.Options = (*TestOption)(nil)

var _ = Describe("options", func() {

	It("skips unknown option", func() {
		set := options.OptionSet{}

		var opt *TestOption
		Expect(set.Get(&opt)).To(BeFalse())
	})

	It("assigns options pointer from set", func() {
		inst := &TestOption{}
		set := options.OptionSet{inst}
		set.Options(inst).(*TestOption).Flag = true

		var opt *TestOption
		Expect(set.Get(&opt)).To(BeTrue())
		Expect(opt.Flag).To(BeTrue())
		Expect(opt).To(BeIdenticalTo(inst))

		Expect(set.Get(&set)).To(BeFalse())
	})

	It("assigns options value from set", func() {
		inst := &TestOption{}
		set := options.OptionSet{inst}

		set.Options(inst).(*TestOption).Flag = true

		var opt TestOption
		Expect(set.Get(&opt)).To(BeTrue())
		Expect(opt.Flag).To(BeTrue())

		opt.Flag = false
		Expect(inst.Flag).To(BeTrue())
	})
})
