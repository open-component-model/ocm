package httpclient_test

import (
	"net/http"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	ocicpi "ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/utils/httpclient"
)

func dur(s string) *ocicpi.Duration {
	d := ocicpi.Duration(s)
	return &d
}

func TestTransport(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Transport Test Suite")
}

var defaultTransport = http.DefaultTransport.(*http.Transport)

var _ = Describe("NewTransport", func() {
	Context("when no config is provided", func() {
		It("preserves http.DefaultTransport values", func() {
			tr, err := httpclient.NewTransport(nil)
			Expect(err).To(Succeed())
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
			tr, err := httpclient.NewTransport(&ocicpi.HTTPSettings{})
			Expect(err).To(Succeed())
			Expect(tr.TLSHandshakeTimeout).To(Equal(defaultTransport.TLSHandshakeTimeout))
			Expect(tr.IdleConnTimeout).To(Equal(defaultTransport.IdleConnTimeout))
			Expect(tr.ResponseHeaderTimeout).To(Equal(defaultTransport.ResponseHeaderTimeout))
			Expect(tr.ExpectContinueTimeout).To(Equal(defaultTransport.ExpectContinueTimeout))
			Expect(tr.MaxIdleConns).To(Equal(defaultTransport.MaxIdleConns))
			Expect(tr.ForceAttemptHTTP2).To(Equal(defaultTransport.ForceAttemptHTTP2))
		})
	})

	Context("when individual fields are set", func() {
		It("overrides TLSHandshakeTimeout only", func() {
			tr, err := httpclient.NewTransport(&ocicpi.HTTPSettings{
				TLSHandshakeTimeout: dur("5s"),
			})
			Expect(err).To(Succeed())
			Expect(tr.TLSHandshakeTimeout).To(Equal(5 * time.Second))
			Expect(tr.IdleConnTimeout).To(Equal(defaultTransport.IdleConnTimeout))
		})

		It("overrides IdleConnTimeout only", func() {
			tr, err := httpclient.NewTransport(&ocicpi.HTTPSettings{
				IdleConnTimeout: dur("120s"),
			})
			Expect(err).To(Succeed())
			Expect(tr.IdleConnTimeout).To(Equal(120 * time.Second))
			Expect(tr.TLSHandshakeTimeout).To(Equal(defaultTransport.TLSHandshakeTimeout))
		})

		It("overrides ResponseHeaderTimeout only", func() {
			tr, err := httpclient.NewTransport(&ocicpi.HTTPSettings{
				ResponseHeaderTimeout: dur("20s"),
			})
			Expect(err).To(Succeed())
			Expect(tr.ResponseHeaderTimeout).To(Equal(20 * time.Second))
			Expect(tr.TLSHandshakeTimeout).To(Equal(defaultTransport.TLSHandshakeTimeout))
		})

		It("replaces DialContext when TCPDialTimeout is set", func() {
			tr, err := httpclient.NewTransport(&ocicpi.HTTPSettings{
				TCPDialTimeout: dur("15s"),
			})
			Expect(err).To(Succeed())
			Expect(tr.DialContext).NotTo(BeNil())
			Expect(tr.TLSHandshakeTimeout).To(Equal(defaultTransport.TLSHandshakeTimeout))
		})
	})

	Context("when all fields are set", func() {
		It("applies all values and preserves non-timeout defaults", func() {
			tr, err := httpclient.NewTransport(&ocicpi.HTTPSettings{
				TCPDialTimeout:        dur("1s"),
				TCPKeepAlive:          dur("2s"),
				TLSHandshakeTimeout:   dur("3s"),
				ResponseHeaderTimeout: dur("4s"),
				IdleConnTimeout:       dur("5s"),
			})
			Expect(err).To(Succeed())
			Expect(tr.TLSHandshakeTimeout).To(Equal(3 * time.Second))
			Expect(tr.ResponseHeaderTimeout).To(Equal(4 * time.Second))
			Expect(tr.IdleConnTimeout).To(Equal(5 * time.Second))
			Expect(tr.ExpectContinueTimeout).To(Equal(defaultTransport.ExpectContinueTimeout))
			Expect(tr.MaxIdleConns).To(Equal(defaultTransport.MaxIdleConns))
			Expect(tr.ForceAttemptHTTP2).To(Equal(defaultTransport.ForceAttemptHTTP2))
		})
	})
})
