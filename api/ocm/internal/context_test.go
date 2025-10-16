package internal_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/ocm"
)

var _ = Describe("OCM Context Test Environment", func() {
	Context("OCM Context", func() {
		It("", func() {
			ctx := ocm.New(datacontext.MODE_EXTENDED)
			Expect(ctx.AttributesContext()).To(BeIdenticalTo(ctx.OCIContext().AttributesContext()))
		})
	})
})
