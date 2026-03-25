package httpclient_test

import (
	"net/http"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/datacontext/attrs/httpcfgattr"
	"ocm.software/ocm/api/utils/httpclient"
)

func TestTransport(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Transport Test Suite")
}

var defaultTransport = http.DefaultTransport.(*http.Transport)

var _ = Describe("NewTransport", func() {
	Context("when no config is provided", func() {
		It("preserves http.DefaultTransport values", func() {
			tr := httpclient.NewTransport(nil)
			Expect(tr.TLSHandshakeTimeout).To(Equal(defaultTransport.TLSHandshakeTimeout))
			Expect(tr.IdleConnTimeout).To(Equal(defaultTransport.IdleConnTimeout))
			Expect(tr.ResponseHeaderTimeout).To(Equal(defaultTransport.ResponseHeaderTimeout))
			Expect(tr.ExpectContinueTimeout).To(Equal(defaultTransport.ExpectContinueTimeout))
			Expect(tr.MaxIdleConns).To(Equal(defaultTransport.MaxIdleConns))
			Expect(tr.ForceAttemptHTTP2).To(Equal(defaultTransport.ForceAttemptHTTP2))
		})
	})

	Context("when config has all nil fields", func() {
		It("preserves http.DefaultTransport values", func() {
			tr := httpclient.NewTransport(&httpcfgattr.HTTPSettings{})
			Expect(tr.TLSHandshakeTimeout).To(Equal(defaultTransport.TLSHandshakeTimeout))
			Expect(tr.IdleConnTimeout).To(Equal(defaultTransport.IdleConnTimeout))
			Expect(tr.ResponseHeaderTimeout).To(Equal(defaultTransport.ResponseHeaderTimeout))
		})
	})

	Context("when default attribute settings are used (no config)", func() {
		It("preserves http.DefaultTransport values", func() {
			attr := &httpcfgattr.Attribute{}
			tr := httpclient.NewTransport(attr.GetHTTPSettings())
			Expect(tr.TLSHandshakeTimeout).To(Equal(defaultTransport.TLSHandshakeTimeout))
			Expect(tr.IdleConnTimeout).To(Equal(defaultTransport.IdleConnTimeout))
			Expect(tr.ResponseHeaderTimeout).To(Equal(defaultTransport.ResponseHeaderTimeout))
			Expect(tr.ExpectContinueTimeout).To(Equal(defaultTransport.ExpectContinueTimeout))
		})
	})

	Context("when individual fields are set", func() {
		It("overrides TLSHandshakeTimeout only", func() {
			tr := httpclient.NewTransport(&httpcfgattr.HTTPSettings{
				TLSHandshakeTimeout: httpcfgattr.NewDuration(5 * time.Second),
			})
			Expect(tr.TLSHandshakeTimeout).To(Equal(5 * time.Second))
			Expect(tr.IdleConnTimeout).To(Equal(defaultTransport.IdleConnTimeout))
		})

		It("overrides IdleConnTimeout only", func() {
			tr := httpclient.NewTransport(&httpcfgattr.HTTPSettings{
				IdleConnTimeout: httpcfgattr.NewDuration(120 * time.Second),
			})
			Expect(tr.IdleConnTimeout).To(Equal(120 * time.Second))
			Expect(tr.TLSHandshakeTimeout).To(Equal(defaultTransport.TLSHandshakeTimeout))
		})

		It("overrides ResponseHeaderTimeout only", func() {
			tr := httpclient.NewTransport(&httpcfgattr.HTTPSettings{
				ResponseHeaderTimeout: httpcfgattr.NewDuration(20 * time.Second),
			})
			Expect(tr.ResponseHeaderTimeout).To(Equal(20 * time.Second))
			Expect(tr.TLSHandshakeTimeout).To(Equal(defaultTransport.TLSHandshakeTimeout))
		})

		It("replaces DialContext when TCPDialTimeout is set", func() {
			tr := httpclient.NewTransport(&httpcfgattr.HTTPSettings{
				TCPDialTimeout: httpcfgattr.NewDuration(15 * time.Second),
			})
			Expect(tr.DialContext).NotTo(BeNil())
			Expect(tr.TLSHandshakeTimeout).To(Equal(defaultTransport.TLSHandshakeTimeout))
		})
	})

	Context("when all fields are set", func() {
		It("applies all values and preserves non-timeout defaults", func() {
			tr := httpclient.NewTransport(&httpcfgattr.HTTPSettings{
				TCPDialTimeout:        httpcfgattr.NewDuration(1 * time.Second),
				TCPKeepAlive:          httpcfgattr.NewDuration(2 * time.Second),
				TLSHandshakeTimeout:   httpcfgattr.NewDuration(3 * time.Second),
				ResponseHeaderTimeout: httpcfgattr.NewDuration(4 * time.Second),
				IdleConnTimeout:       httpcfgattr.NewDuration(5 * time.Second),
			})
			Expect(tr.TLSHandshakeTimeout).To(Equal(3 * time.Second))
			Expect(tr.ResponseHeaderTimeout).To(Equal(4 * time.Second))
			Expect(tr.IdleConnTimeout).To(Equal(5 * time.Second))
			Expect(tr.ExpectContinueTimeout).To(Equal(defaultTransport.ExpectContinueTimeout))
			Expect(tr.MaxIdleConns).To(Equal(defaultTransport.MaxIdleConns))
			Expect(tr.ForceAttemptHTTP2).To(Equal(defaultTransport.ForceAttemptHTTP2))
		})
	})
})
