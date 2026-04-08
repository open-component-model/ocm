package config_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/oci/config"
	"ocm.software/ocm/api/oci/cpi"
)

func dur(s string) *cpi.Duration {
	d := cpi.Duration(s)
	return &d
}

func MustGetHTTPSettings(ctx cpi.Context) cpi.HTTPSettings {
	g, err := ctx.GetHTTPSettings()
	ExpectWithOffset(1, err).To(Succeed())
	return g
}

var _ = Describe("http config", func() {
	Context("apply", func() {
		It("applies timeout via ApplyTo", func() {
			ctx := cpi.New()
			cfg := &config.HTTPConfig{HTTPSettings: cpi.HTTPSettings{Timeout: dur("5m")}}

			Expect(cfg.ApplyTo(ctx.ConfigContext(), ctx)).To(Succeed())
			g := MustGetHTTPSettings(ctx)
			Expect(g.Timeout.TimeDuration()).To(HaveValue(Equal(5 * time.Minute)))
		})

		It("applies via config context", func() {
			ctx := cpi.New()
			cfg := &config.HTTPConfig{HTTPSettings: cpi.HTTPSettings{Timeout: dur("30s")}}

			Expect(ctx.ConfigContext().ApplyConfig(cfg, "programmatic")).To(Succeed())
			g := MustGetHTTPSettings(ctx)
			Expect(g.Timeout.TimeDuration()).To(HaveValue(Equal(30 * time.Second)))
		})

		It("parses all fields from JSON config", func() {
			ctx := cpi.New()
			raw := []byte(`{"type":"http.config.ocm.software/v1alpha1","timeout":"10s","tcpDialTimeout":"15s","tcpKeepAlive":"20s","tlsHandshakeTimeout":"5s","responseHeaderTimeout":"8s","idleConnTimeout":"45s"}`)
			cfg, err := ctx.ConfigContext().GetConfigForData(raw, nil)
			Expect(err).To(Succeed())
			Expect(ctx.ConfigContext().ApplyConfig(cfg, "config file")).To(Succeed())

			g := MustGetHTTPSettings(ctx)
			Expect(g.Timeout.TimeDuration()).To(HaveValue(Equal(10 * time.Second)))
			Expect(g.TCPDialTimeout.TimeDuration()).To(HaveValue(Equal(15 * time.Second)))
			Expect(g.TCPKeepAlive.TimeDuration()).To(HaveValue(Equal(20 * time.Second)))
			Expect(g.TLSHandshakeTimeout.TimeDuration()).To(HaveValue(Equal(5 * time.Second)))
			Expect(g.ResponseHeaderTimeout.TimeDuration()).To(HaveValue(Equal(8 * time.Second)))
			Expect(g.IdleConnTimeout.TimeDuration()).To(HaveValue(Equal(45 * time.Second)))
		})

		It("successive ApplyConfig overrides only non-nil fields", func() {
			ctx := cpi.New()

			first := &config.HTTPConfig{
				HTTPSettings: cpi.HTTPSettings{
					TCPDialTimeout: dur("15s"),
				},
			}
			Expect(first.ApplyTo(ctx.ConfigContext(), ctx)).To(Succeed())

			second := &config.HTTPConfig{HTTPSettings: cpi.HTTPSettings{Timeout: dur("1m")}}
			Expect(second.ApplyTo(ctx.ConfigContext(), ctx)).To(Succeed())

			g := MustGetHTTPSettings(ctx)
			Expect(g.Timeout.TimeDuration()).To(HaveValue(Equal(1 * time.Minute)))
			Expect(g.TCPDialTimeout.TimeDuration()).To(HaveValue(Equal(15 * time.Second)))
		})

		It("applies via generic config wrapper", func() {
			ctx := cpi.New()
			raw := []byte(`
type: generic.config.ocm.software/v1
configurations:
  - type: http.config.ocm.software
    timeout: "10s"
    tcpDialTimeout: "15s"
    tcpKeepAlive: "20s"
    tlsHandshakeTimeout: "5s"
    responseHeaderTimeout: "8s"
    idleConnTimeout: "45s"
`)
			cfg, err := ctx.ConfigContext().GetConfigForData(raw, nil)
			Expect(err).To(Succeed())
			Expect(ctx.ConfigContext().ApplyConfig(cfg, "config file")).To(Succeed())

			g := MustGetHTTPSettings(ctx)
			Expect(g.Timeout.TimeDuration()).To(HaveValue(Equal(10 * time.Second)))
			Expect(g.TCPDialTimeout.TimeDuration()).To(HaveValue(Equal(15 * time.Second)))
			Expect(g.TCPKeepAlive.TimeDuration()).To(HaveValue(Equal(20 * time.Second)))
			Expect(g.TLSHandshakeTimeout.TimeDuration()).To(HaveValue(Equal(5 * time.Second)))
			Expect(g.ResponseHeaderTimeout.TimeDuration()).To(HaveValue(Equal(8 * time.Second)))
			Expect(g.IdleConnTimeout.TimeDuration()).To(HaveValue(Equal(45 * time.Second)))
		})

		It("rejects invalid duration string", func() {
			ctx := cpi.New()
			raw := []byte(`{"type":"http.config.ocm.software/v1alpha1","timeout":"notaduration"}`)
			_, err := ctx.ConfigContext().GetConfigForData(raw, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid duration: notaduration"))
		})

		It("rejects negative duration", func() {
			ctx := cpi.New()
			raw := []byte(`{"type":"http.config.ocm.software/v1alpha1","timeout":"-5m"}`)
			_, err := ctx.ConfigContext().GetConfigForData(raw, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("negative duration not allowed: -5m"))
		})

		It("default settings are nil", func() {
			g := MustGetHTTPSettings(cpi.New())
			Expect(g.Timeout).To(BeNil())
			Expect(g.TCPDialTimeout).To(BeNil())
			Expect(g.TCPKeepAlive).To(BeNil())
			Expect(g.TLSHandshakeTimeout).To(BeNil())
			Expect(g.ResponseHeaderTimeout).To(BeNil())
			Expect(g.IdleConnTimeout).To(BeNil())
		})

		It("nil timeout returns nil not zero", func() {
			g := MustGetHTTPSettings(cpi.New())
			timeout, err := g.Timeout.TimeDuration()
			Expect(err).To(Succeed())
			Expect(timeout).To(BeNil())
		})
	})
})
