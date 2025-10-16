package spiff_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/spiff/spiffing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/utils/spiff"
)

var _ = Describe("spiff", func() {
	It("calls spiff", func() {
		res := Must(spiff.CascadeWith(spiff.TemplateData("test", []byte("((  \"alice+\" \"bob\" ))")), spiff.Mode(spiffing.MODE_PRIVATE)))
		Expect(string(res)).To(Equal("alice+bob\n"))
	})
})
