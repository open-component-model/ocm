package plugindirattr_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/config"
	"ocm.software/ocm/api/datacontext"
	me "ocm.software/ocm/api/ocm/extensions/attrs/plugindirattr"
)

var _ = Describe("attribute", func() {
	var ctx config.Context

	attr := "___test___"

	BeforeEach(func() {
		ctx = config.WithSharedAttributes(datacontext.New(nil)).New()
	})

	It("local setting", func() {
		Expect(me.Get(ctx)).NotTo(Equal(attr))
		Expect(me.Set(ctx, attr)).To(Succeed())
		Expect(me.Get(ctx)).To(BeIdenticalTo(attr))
	})
})
