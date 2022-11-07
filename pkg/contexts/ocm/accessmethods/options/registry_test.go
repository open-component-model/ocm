// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package options

import (
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
)

var _ = Describe("registry", func() {
	var reg Registry

	BeforeEach(func() {
		reg = New()
	})

	It("sets and retrieves type", func() {
		reg.RegisterType("string", flagsets.NewStringOptionType)

		t := reg.GetType("string")
		Expect(t).NotTo(BeNil())

		o := Must(reg.CreateOption("string", "test", "some test"))
		Expect(o.Name()).To(Equal("test"))
		Expect(o.Description()).To(Equal("some test"))
		Expect(reflect.TypeOf(o)).To(Equal(reflect.TypeOf(flagsets.NewStringOptionType("", ""))))
	})

	It("sets and retrieves option", func() {
		reg.RegisterOption(HostnameOption)

		t := reg.GetOption(HostnameOption.Name())
		Expect(t).NotTo(BeNil())
	})

	It("creates merges a new type", func() {
		reg.RegisterType("string", flagsets.NewStringOptionType)
		reg.RegisterOption(HostnameOption)

		o := Must(reg.CreateOption("string", HostnameOption.Name(), "some test"))
		Expect(o).To(BeIdenticalTo(HostnameOption))
	})

	It("fails creating existing", func() {
		reg.RegisterType("string", flagsets.NewStringOptionType)
		reg.RegisterType("int", flagsets.NewIntOptionType)
		reg.RegisterOption(HostnameOption)

		_, err := reg.CreateOption("int", HostnameOption.Name(), "some test")
		MustFailWithMessage(err, "option \"accessHostname\" already exists")
	})
})
