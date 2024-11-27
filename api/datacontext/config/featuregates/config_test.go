package featuregates_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/datacontext/attrs/featuregatesattr"
	"ocm.software/ocm/api/datacontext/config/attrs"
	"ocm.software/ocm/api/datacontext/config/featuregates"

	"ocm.software/ocm/api/config"
	"ocm.software/ocm/api/datacontext"
)

var _ = Describe("feature gates", func() {
	var ctx config.Context

	BeforeEach(func() {
		ctx = config.WithSharedAttributes(datacontext.New(nil)).New()
	})

	Context("applies", func() {
		It("handles default", func() {
			a := featuregatesattr.Get(ctx)

			Expect(a.IsEnabled("test")).To(BeFalse())
			Expect(a.IsEnabled("test", true)).To(BeTrue())
			g := a.GetFeature("test", true)
			Expect(g).NotTo(BeNil())
			Expect(g.Mode).To(Equal(""))
		})

		It("enables feature", func() {
			cfg := featuregates.New()
			cfg.EnableFeature("test", &featuregates.FeatureGate{Mode: "on"})
			ctx.ApplyConfig(cfg, "manual")

			a := featuregatesattr.Get(ctx)

			Expect(a.IsEnabled("test")).To(BeTrue())
			Expect(a.IsEnabled("test", true)).To(BeTrue())
			g := a.GetFeature("test")
			Expect(g).NotTo(BeNil())
			Expect(g.Mode).To(Equal("on"))
		})

		It("disables feature", func() {
			cfg := featuregates.New()
			cfg.DisableFeature("test")
			ctx.ApplyConfig(cfg, "manual")

			a := featuregatesattr.Get(ctx)

			Expect(a.IsEnabled("test")).To(BeFalse())
			Expect(a.IsEnabled("test", true)).To(BeFalse())
		})

		It("handle attribute config", func() {
			cfg := featuregatesattr.New()
			cfg.EnableFeature("test", &featuregates.FeatureGate{Mode: "on"})

			spec := attrs.New()
			Expect(spec.AddAttribute(featuregatesattr.ATTR_KEY, cfg)).To(Succeed())
			Expect(ctx.ApplyConfig(spec, "test")).To(Succeed())

			ctx.ApplyConfig(spec, "manual")

			a := featuregatesattr.Get(ctx)

			Expect(a.IsEnabled("test")).To(BeTrue())
			g := a.GetFeature("test")
			Expect(g).NotTo(BeNil())
			Expect(g.Mode).To(Equal("on"))
		})

	})
})
