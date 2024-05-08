package compositionmodeattr_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/config"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	me "github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/compositionmodeattr"
	"github.com/open-component-model/ocm/pkg/runtime"
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
