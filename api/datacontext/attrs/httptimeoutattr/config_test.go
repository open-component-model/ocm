package httptimeoutattr_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/config"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/datacontext/attrs/httptimeoutattr"
)

var _ = Describe("http config type", func() {
	var ctx config.Context

	BeforeEach(func() {
		ctx = config.WithSharedAttributes(datacontext.New(nil)).New()
	})

	It("applies timeout via config", func() {
		cfg := httptimeoutattr.NewConfig(5 * time.Minute)
		Expect(ctx.ApplyConfig(cfg, "test")).To(Succeed())

		ocmCtx := credentials.WithConfigs(ctx).New()
		Expect(httptimeoutattr.Get(ocmCtx)).To(Equal(5 * time.Minute))
	})

	It("applies timeout to existing context", func() {
		ocmCtx := credentials.WithConfigs(ctx).New()
		cfg := httptimeoutattr.NewConfig(2 * time.Minute)
		Expect(ctx.ApplyConfig(cfg, "test")).To(Succeed())

		Expect(httptimeoutattr.Get(ocmCtx)).To(Equal(2 * time.Minute))
	})

	It("skips zero timeout", func() {
		cfg := httptimeoutattr.NewConfig(0)
		Expect(ctx.ApplyConfig(cfg, "test")).To(Succeed())

		ocmCtx := credentials.WithConfigs(ctx).New()
		Expect(httptimeoutattr.Get(ocmCtx)).To(Equal(httptimeoutattr.DefaultTimeout))
	})

	It("applies timeout from config file with timeout field", func() {
		raw := []byte(`{"type":"http.config.ocm.software/v1alpha1","timeout":"10s"}`)
		cfg, err := ctx.GetConfigForData(raw, nil)
		Expect(err).To(Succeed())
		Expect(ctx.ApplyConfig(cfg, "config file")).To(Succeed())

		ocmCtx := credentials.WithConfigs(ctx).New()
		Expect(httptimeoutattr.Get(ocmCtx)).To(Equal(10 * time.Second))
	})

	It("flag overrides config file timeout", func() {
		// Config file sets timeout to 10s.
		raw := []byte(`{"type":"http.config.ocm.software/v1alpha1","timeout":"10s"}`)
		cfg, err := ctx.GetConfigForData(raw, nil)
		Expect(err).To(Succeed())
		Expect(ctx.ApplyConfig(cfg, "config file")).To(Succeed())

		// CLI flag sets timeout to 2m (applied after config file, same as in Evaluate).
		flagCfg := httptimeoutattr.NewConfig(2 * time.Minute)
		Expect(ctx.ApplyConfig(flagCfg, "cli timeout flag")).To(Succeed())

		ocmCtx := credentials.WithConfigs(ctx).New()
		Expect(httptimeoutattr.Get(ocmCtx)).To(Equal(2 * time.Minute))
	})

	It("keeps default timeout when config file omits timeout field", func() {
		raw := []byte(`{"type":"http.config.ocm.software/v1alpha1"}`)
		cfg, err := ctx.GetConfigForData(raw, nil)
		Expect(err).To(Succeed())
		Expect(ctx.ApplyConfig(cfg, "config file")).To(Succeed())

		ocmCtx := credentials.WithConfigs(ctx).New()
		Expect(httptimeoutattr.Get(ocmCtx)).To(Equal(httptimeoutattr.DefaultTimeout))
	})
})
