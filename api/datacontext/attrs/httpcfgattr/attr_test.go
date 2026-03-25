package httpcfgattr_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/datacontext/attrs/httpcfgattr"
	"ocm.software/ocm/api/utils/runtime"
)

var _ = Describe("httpcfg attribute", func() {
	var ctx datacontext.Context
	attr := httpcfgattr.AttributeType{}
	enc := runtime.DefaultJSONEncoding

	BeforeEach(func() {
		ctx = datacontext.New(nil)
	})

	Context("get and set", func() {
		It("defaults to empty settings with no timeout", func() {
			a := httpcfgattr.Get(ctx)
			Expect(a).NotTo(BeNil())
			Expect(a.GetHTTPSettings().GetTimeout()).To(Equal(time.Duration(0)))
		})

		It("round-trips through encode/decode", func() {
			a := httpcfgattr.Get(ctx)
			cfg := httpcfgattr.NewConfig(30 * time.Second)
			cfg.ApplyToAttribute(a)

			Expect(httpcfgattr.Get(ctx).GetHTTPSettings().GetTimeout()).To(Equal(30 * time.Second))
		})
	})

	Context("encoding", func() {
		It("encodes *Attribute to JSON", func() {
			a := httpcfgattr.Get(ctx)
			cfg := httpcfgattr.NewConfig(30 * time.Second)
			cfg.ApplyToAttribute(a)

			data, err := attr.Encode(a, enc)
			Expect(err).To(Succeed())
			Expect(string(data)).To(ContainSubstring(`"timeout":"30s"`))
		})

		It("rejects non-*Attribute input", func() {
			_, err := attr.Encode("invalid", enc)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("is invalid"))
		})
	})

	Context("decoding", func() {
		It("decodes JSON to *Attribute", func() {
			raw := []byte(`{"type":"http.config.ocm.software/v1alpha1","timeout":"10s"}`)
			val, err := attr.Decode(raw, enc)
			Expect(err).To(Succeed())
			a, ok := val.(*httpcfgattr.Attribute)
			Expect(ok).To(BeTrue())
			Expect(a.GetHTTPSettings().GetTimeout()).To(Equal(10 * time.Second))
		})

		It("rejects invalid JSON", func() {
			_, err := attr.Decode([]byte(`{invalid`), enc)
			Expect(err).To(HaveOccurred())
		})
	})
})
