package flagsets_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"

	"ocm.software/ocm/api/ocm/extensions/accessmethods/options"
)

var _ = Describe("registry", func() {
	var reg flagsets.ConfigOptionTypeRegistry

	BeforeEach(func() {
		reg = flagsets.NewConfigOptionTypeRegistry()
	})

	It("sets and retrieves type", func() {
		reg.RegisterValueType("string", flagsets.NewStringOptionType, "string")

		t := reg.GetValueType("string")
		Expect(t).NotTo(BeNil())

		o := Must(reg.CreateOptionType("string", "test", "some test"))
		Expect(o.GetName()).To(Equal("test"))
		Expect(o.GetDescription()).To(Equal("[*string*] some test"))
	})

	It("sets and retrieves option", func() {
		reg.RegisterOptionType(options.HostnameOption)

		t := reg.GetOptionType(options.HostnameOption.GetName())
		Expect(t).NotTo(BeNil())
	})

	It("creates merges a new type", func() {
		reg.RegisterValueType("string", flagsets.NewStringOptionType, "string")
		reg.RegisterOptionType(options.HostnameOption)

		o := Must(reg.CreateOptionType("string", options.HostnameOption.GetName(), "some test"))
		Expect(o).To(BeIdenticalTo(options.HostnameOption))
	})

	It("fails creating existing", func() {
		reg.RegisterValueType("string", flagsets.NewStringOptionType, "string")
		reg.RegisterValueType("int", flagsets.NewIntOptionType, "int")
		reg.RegisterOptionType(options.HostnameOption)

		_, err := reg.CreateOptionType("int", options.HostnameOption.GetName(), "some test")
		MustFailWithMessage(err, "option \"accessHostname\" already exists")
	})
})
