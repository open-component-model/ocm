// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

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
