package utils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	me "github.com/open-component-model/ocm/pkg/utils"
)

var _ = Describe("package tests", func() {
	It("go module name", func() {
		mod := Must(me.GetModuleName())
		Expect(mod).To(Equal("github.com/open-component-model/ocm"))
	})
})
