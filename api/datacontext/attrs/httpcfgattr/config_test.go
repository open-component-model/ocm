package httpcfgattr_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/config"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/datacontext/attrs/httpcfgattr"
)

var _ = Describe("http config type", func() {
	var ctx config.Context

	BeforeEach(func() {
		ctx = config.WithSharedAttributes(datacontext.New(nil)).New()
	})

	It("applies timeout via ApplyConfig", func() {
		cfg := httpcfgattr.NewConfig(5 * time.Minute)
		Expect(ctx.ApplyConfig(cfg, "test")).To(Succeed())

		ocmCtx := credentials.WithConfigs(ctx).New()
		Expect(httpcfgattr.Get(ocmCtx).GetHTTPSettings().GetTimeout()).To(Equal(5 * time.Minute))
	})

	It("parses all fields from JSON config", func() {
		raw := []byte(`{"type":"http.config.ocm.software/v1alpha1","timeout":"10s","tcpDialTimeout":"15s","tcpKeepAlive":"20s","tlsHandshakeTimeout":"5s","responseHeaderTimeout":"8s","idleConnTimeout":"45s"}`)
		cfg, err := ctx.GetConfigForData(raw, nil)
		Expect(err).To(Succeed())
		Expect(ctx.ApplyConfig(cfg, "config file")).To(Succeed())

		ocmCtx := credentials.WithConfigs(ctx).New()
		g := httpcfgattr.Get(ocmCtx).GetHTTPSettings()
		Expect(g.GetTimeout()).To(Equal(10 * time.Second))
		Expect(g.TCPDialTimeout.TimeDuration()).To(Equal(15 * time.Second))
		Expect(g.TCPKeepAlive.TimeDuration()).To(Equal(20 * time.Second))
		Expect(g.TLSHandshakeTimeout.TimeDuration()).To(Equal(5 * time.Second))
		Expect(g.ResponseHeaderTimeout.TimeDuration()).To(Equal(8 * time.Second))
		Expect(g.IdleConnTimeout.TimeDuration()).To(Equal(45 * time.Second))
	})

	It("parses all fields from YAML config", func() {
		raw := []byte(`
type: http.config.ocm.software/v1alpha1
timeout: "0s"
tcpDialTimeout: "30s"
tcpKeepAlive: "30s"
tlsHandshakeTimeout: "10s"
responseHeaderTimeout: "10s"
idleConnTimeout: "90s"
`)
		cfg, err := ctx.GetConfigForData(raw, nil)
		Expect(err).To(Succeed())
		Expect(ctx.ApplyConfig(cfg, "yaml config file")).To(Succeed())

		ocmCtx := credentials.WithConfigs(ctx).New()
		g := httpcfgattr.Get(ocmCtx).GetHTTPSettings()
		Expect(g.GetTimeout()).To(Equal(time.Duration(0)))
		Expect(g.TCPDialTimeout.TimeDuration()).To(Equal(30 * time.Second))
		Expect(g.TCPKeepAlive.TimeDuration()).To(Equal(30 * time.Second))
		Expect(g.TLSHandshakeTimeout.TimeDuration()).To(Equal(10 * time.Second))
		Expect(g.ResponseHeaderTimeout.TimeDuration()).To(Equal(10 * time.Second))
		Expect(g.IdleConnTimeout.TimeDuration()).To(Equal(90 * time.Second))
	})

	It("successive ApplyConfig overrides earlier values", func() {
		raw := []byte(`{"type":"http.config.ocm.software/v1alpha1","timeout":"10s"}`)
		cfg, err := ctx.GetConfigForData(raw, nil)
		Expect(err).To(Succeed())
		Expect(ctx.ApplyConfig(cfg, "config file")).To(Succeed())

		override := httpcfgattr.NewConfig(2 * time.Minute)
		Expect(ctx.ApplyConfig(override, "cli")).To(Succeed())

		ocmCtx := credentials.WithConfigs(ctx).New()
		Expect(httpcfgattr.Get(ocmCtx).GetHTTPSettings().GetTimeout()).To(Equal(2 * time.Minute))
	})

	It("rejects invalid duration string", func() {
		raw := []byte(`{"type":"http.config.ocm.software/v1alpha1","timeout":"notaduration"}`)
		_, err := ctx.GetConfigForData(raw, nil)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("invalid duration: notaduration"))
	})

	It("returns nil fields when config omits timeout fields", func() {
		raw := []byte(`{"type":"http.config.ocm.software/v1alpha1"}`)
		cfg, err := ctx.GetConfigForData(raw, nil)
		Expect(err).To(Succeed())
		Expect(ctx.ApplyConfig(cfg, "config file")).To(Succeed())

		ocmCtx := credentials.WithConfigs(ctx).New()
		g := httpcfgattr.Get(ocmCtx).GetHTTPSettings()
		Expect(g.Timeout).To(BeNil())
		Expect(g.TCPDialTimeout).To(BeNil())
	})

	It("default attribute returns empty settings", func() {
		ocmCtx := credentials.WithConfigs(ctx).New()
		g := httpcfgattr.Get(ocmCtx).GetHTTPSettings()
		Expect(g.Timeout).To(BeNil())
		Expect(g.TCPDialTimeout).To(BeNil())
		Expect(g.TCPKeepAlive).To(BeNil())
		Expect(g.TLSHandshakeTimeout).To(BeNil())
		Expect(g.ResponseHeaderTimeout).To(BeNil())
		Expect(g.IdleConnTimeout).To(BeNil())
	})

	It("ApplyToAttribute overrides only non-nil fields", func() {
		attr := &httpcfgattr.Attribute{}
		first := &httpcfgattr.Config{
			HTTPSettings: httpcfgattr.HTTPSettings{
				TCPDialTimeout: httpcfgattr.NewDuration(15 * time.Second),
			},
		}
		second := &httpcfgattr.Config{
			HTTPSettings: httpcfgattr.HTTPSettings{
				Timeout: httpcfgattr.NewDuration(1 * time.Minute),
			},
		}
		first.ApplyToAttribute(attr)
		second.ApplyToAttribute(attr)
		g := attr.GetHTTPSettings()
		Expect(g.GetTimeout()).To(Equal(1 * time.Minute))
		Expect(g.TCPDialTimeout.TimeDuration()).To(Equal(15 * time.Second))
	})

	It("ApplyToAttribute with empty config leaves attribute unchanged", func() {
		attr := &httpcfgattr.Attribute{}
		first := &httpcfgattr.Config{
			HTTPSettings: httpcfgattr.HTTPSettings{
				TCPDialTimeout: httpcfgattr.NewDuration(5 * time.Second),
			},
		}
		first.ApplyToAttribute(attr)

		empty := &httpcfgattr.Config{}
		empty.ApplyToAttribute(attr)

		g := attr.GetHTTPSettings()
		Expect(g.TCPDialTimeout.TimeDuration()).To(Equal(5 * time.Second))
		Expect(g.Timeout).To(BeNil())
	})
})
