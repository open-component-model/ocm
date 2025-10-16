package ocm_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	me "ocm.software/ocm/api/ocm"
)

var _ = Describe("area test", func() {
	It("binds to Go context", func() {
		ctx := context.Background()

		mine := me.New()
		nctx := mine.BindTo(ctx)

		me.FromContext(nctx)
		Expect(me.FromContext(nctx)).To(BeIdenticalTo(mine))
	})
})
