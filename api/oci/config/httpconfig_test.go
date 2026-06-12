package config_test

import (
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/oci/config"
	"ocm.software/ocm/api/oci/cpi"
)

func dur(s string) *cpi.Duration {
	td, err := time.ParseDuration(s)
	if err != nil {
		panic(err)
	}
	d := cpi.Duration(td)
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
			Expect(time.Duration(*g.Timeout)).To(Equal(5 * time.Minute))
		})

		It("applies via config context", func() {
			ctx := cpi.New()
			cfg := &config.HTTPConfig{HTTPSettings: cpi.HTTPSettings{Timeout: dur("30s")}}

			Expect(ctx.ConfigContext().ApplyConfig(cfg, "programmatic")).To(Succeed())
			g := MustGetHTTPSettings(ctx)
			Expect(time.Duration(*g.Timeout)).To(Equal(30 * time.Second))
		})

		It("parses all fields from JSON config", func() {
			ctx := cpi.New()
			raw := []byte(`{"type":"http.config.ocm.software/v1alpha1","timeout":"10s","tcpDialTimeout":"15s","tcpKeepAlive":"20s","tlsHandshakeTimeout":"5s","responseHeaderTimeout":"8s","idleConnTimeout":"45s"}`)
			cfg, err := ctx.ConfigContext().GetConfigForData(raw, nil)
			Expect(err).To(Succeed())
			Expect(ctx.ConfigContext().ApplyConfig(cfg, "config file")).To(Succeed())

			g := MustGetHTTPSettings(ctx)
			Expect(time.Duration(*g.Timeout)).To(Equal(10 * time.Second))
			Expect(time.Duration(*g.TCPDialTimeout)).To(Equal(15 * time.Second))
			Expect(time.Duration(*g.TCPKeepAlive)).To(Equal(20 * time.Second))
			Expect(time.Duration(*g.TLSHandshakeTimeout)).To(Equal(5 * time.Second))
			Expect(time.Duration(*g.ResponseHeaderTimeout)).To(Equal(8 * time.Second))
			Expect(time.Duration(*g.IdleConnTimeout)).To(Equal(45 * time.Second))
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
			Expect(time.Duration(*g.Timeout)).To(Equal(1 * time.Minute))
			Expect(time.Duration(*g.TCPDialTimeout)).To(Equal(15 * time.Second))
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
			Expect(time.Duration(*g.Timeout)).To(Equal(10 * time.Second))
			Expect(time.Duration(*g.TCPDialTimeout)).To(Equal(15 * time.Second))
			Expect(time.Duration(*g.TCPKeepAlive)).To(Equal(20 * time.Second))
			Expect(time.Duration(*g.TLSHandshakeTimeout)).To(Equal(5 * time.Second))
			Expect(time.Duration(*g.ResponseHeaderTimeout)).To(Equal(8 * time.Second))
			Expect(time.Duration(*g.IdleConnTimeout)).To(Equal(45 * time.Second))
		})

		DescribeTable("rejects invalid duration values on unmarshal",
			func(jsonValue string, expectedErr string) {
				ctx := cpi.New()
				raw := []byte(`{"type":"http.config.ocm.software/v1alpha1","timeout":` + jsonValue + `}`)
				_, err := ctx.ConfigContext().GetConfigForData(raw, nil)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(expectedErr))
			},
			Entry("garbage string", `"notaduration"`, "expected a Go duration string"),
			Entry("number instead of string", `42`, "expected a Go duration string"),
			Entry("boolean instead of string", `true`, "expected a Go duration string"),
			Entry("empty string", `""`, "expected a Go duration string"),
		)

		DescribeTable("rejects negative duration for timeout fields",
			func(field string, cfg *config.HTTPConfig) {
				ctx := cpi.New()
				Expect(cfg.ApplyTo(ctx.ConfigContext(), ctx)).To(MatchError(
					ContainSubstring("invalid value for " + field),
				))
			},
			Entry("timeout -5m", "timeout",
				&config.HTTPConfig{HTTPSettings: cpi.HTTPSettings{Timeout: dur("-5m")}}),
			Entry("tcpDialTimeout -10s", "tcpDialTimeout",
				&config.HTTPConfig{HTTPSettings: cpi.HTTPSettings{TCPDialTimeout: dur("-10s")}}),
			Entry("tlsHandshakeTimeout -10h5m", "tlsHandshakeTimeout",
				&config.HTTPConfig{HTTPSettings: cpi.HTTPSettings{TLSHandshakeTimeout: dur("-10h5m")}}),
			Entry("responseHeaderTimeout -1s", "responseHeaderTimeout",
				&config.HTTPConfig{HTTPSettings: cpi.HTTPSettings{ResponseHeaderTimeout: dur("-1s")}}),
			Entry("idleConnTimeout -30s", "idleConnTimeout",
				&config.HTTPConfig{HTTPSettings: cpi.HTTPSettings{IdleConnTimeout: dur("-30s")}}),
		)

		It("allows negative tcpKeepAlive to disable keep-alive probes", func() {
			ctx := cpi.New()
			cfg := &config.HTTPConfig{HTTPSettings: cpi.HTTPSettings{TCPKeepAlive: dur("-1s")}}
			Expect(cfg.ApplyTo(ctx.ConfigContext(), ctx)).To(Succeed())
			g := MustGetHTTPSettings(ctx)
			Expect(time.Duration(*g.TCPKeepAlive)).To(Equal(-1 * time.Second))
		})

		DescribeTable("accepts compound duration like 1h5s",
			func(cfg *config.HTTPConfig) {
				ctx := cpi.New()
				Expect(cfg.ApplyTo(ctx.ConfigContext(), ctx)).To(Succeed())
			},
			Entry("timeout", &config.HTTPConfig{HTTPSettings: cpi.HTTPSettings{Timeout: dur("1h5s")}}),
			Entry("tcpDialTimeout", &config.HTTPConfig{HTTPSettings: cpi.HTTPSettings{TCPDialTimeout: dur("1h5s")}}),
			Entry("tcpKeepAlive", &config.HTTPConfig{HTTPSettings: cpi.HTTPSettings{TCPKeepAlive: dur("1h5s")}}),
			Entry("tlsHandshakeTimeout", &config.HTTPConfig{HTTPSettings: cpi.HTTPSettings{TLSHandshakeTimeout: dur("1h5s")}}),
			Entry("responseHeaderTimeout", &config.HTTPConfig{HTTPSettings: cpi.HTTPSettings{ResponseHeaderTimeout: dur("1h5s")}}),
			Entry("idleConnTimeout", &config.HTTPConfig{HTTPSettings: cpi.HTTPSettings{IdleConnTimeout: dur("1h5s")}}),
		)

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
			Expect(g.Timeout).To(BeNil())
		})

		It("round-trips Duration through MarshalJSON and UnmarshalJSON", func() {
			original := cpi.HTTPSettings{
				Timeout:        dur("5m30s"),
				TCPDialTimeout: dur("15s"),
			}
			data, err := json.Marshal(original)
			Expect(err).To(Succeed())

			var restored cpi.HTTPSettings
			Expect(json.Unmarshal(data, &restored)).To(Succeed())
			Expect(time.Duration(*restored.Timeout)).To(Equal(5*time.Minute + 30*time.Second))
			Expect(time.Duration(*restored.TCPDialTimeout)).To(Equal(15 * time.Second))
			Expect(restored.TCPKeepAlive).To(BeNil())
		})
	})
})
