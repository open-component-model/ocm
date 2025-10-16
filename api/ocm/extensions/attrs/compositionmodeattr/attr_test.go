package compositionmodeattr_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/config"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/ocm"
	me "ocm.software/ocm/api/ocm/extensions/attrs/compositionmodeattr"
	"ocm.software/ocm/api/utils/runtime"
)

var _ = Describe("attribute", func() {
	var ctx ocm.Context
	var cfgctx config.Context

	BeforeEach(func() {
		cfgctx = config.WithSharedAttributes(datacontext.New(nil)).New()
		credctx := credentials.WithConfigs(cfgctx).New()
		ocictx := oci.WithCredentials(credctx).New()
		ctx = ocm.WithOCIRepositories(ocictx).New()
	})
	It("local setting", func() {
		Expect(me.Get(ctx)).To(Equal(me.UseCompositionMode))
		Expect(me.Set(ctx, true)).To(Succeed())
		Expect(me.Get(ctx)).To(BeTrue())
		Expect(me.Set(ctx, false)).To(Succeed())
		Expect(me.Get(ctx)).To(BeFalse())
	})

	It("global setting", func() {
		Expect(me.Get(cfgctx)).To(Equal(me.UseCompositionMode))
		Expect(me.Set(cfgctx, true)).To(Succeed())
		Expect(me.Get(cfgctx)).To(BeTrue())
		Expect(me.Set(cfgctx, false)).To(Succeed())
		Expect(me.Get(cfgctx)).To(BeFalse())
	})

	It("parses string", func() {
		Expect(me.AttributeType{}.Decode([]byte("true"), runtime.DefaultJSONEncoding)).To(BeTrue())
	})
})
