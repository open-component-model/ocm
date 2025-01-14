package logging_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/logging"
	logcfg "github.com/mandelsoft/logging/config"
	"github.com/tonglil/buflogr"

	local "ocm.software/ocm/api/utils/logging"
)

var _ = Describe("RoundTripper", func() {
	var buf bytes.Buffer
	var ctx *local.StaticContext
	var roundTripper http.RoundTripper
	var server *httptest.Server

	BeforeEach(func() {
		buf.Reset()
		local.SetContext(logging.NewDefault())
		ctx = local.Context()
		ctx.SetBaseLogger(buflogr.NewWithBuffer(&buf))
	})

	AfterEach(func() {
		if server != nil {
			server.Close()
		}
	})

	It("redacts Authorization header", func() {
		r := logcfg.ConditionalRule("trace")
		cfg := &logcfg.Config{
			Rules: []logcfg.Rule{r},
		}
		Expect(ctx.Configure(cfg)).To(Succeed())

		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		roundTripper = local.NewRoundTripper(http.DefaultTransport, ctx.Logger())
		client := &http.Client{Transport: roundTripper}

		req, err := http.NewRequest("GET", server.URL, nil)
		Expect(err).NotTo(HaveOccurred())
		req.Header.Set("Authorization", "this should be redacted")

		_, err = client.Do(req)
		Expect(err).NotTo(HaveOccurred())

		Expect(buf.String()).To(ContainSubstring("roundtrip"))
		Expect(buf.String()).To(ContainSubstring("url"))
		Expect(buf.String()).To(ContainSubstring("method"))
		Expect(buf.String()).To(ContainSubstring("header"))
		Expect(buf.String()).To(ContainSubstring("***"))
		Expect(buf.String()).NotTo(ContainSubstring("this should be redacted"))
	})
})
