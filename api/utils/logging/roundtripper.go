package logging

import (
	"net/http"

	"github.com/mandelsoft/logging"
)

func NewRoundTripper(rt http.RoundTripper, logger logging.Logger) *RoundTripper {
	return &RoundTripper{
		logger:       logger,
		RoundTripper: rt,
	}
}

// RoundTripper is a http.RoundTripper that logs requests.
type RoundTripper struct {
	logger logging.Logger
	http.RoundTripper
}

func (t *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Redact sensitive headers to make sure they don't get logged at any point.
	header := req.Header.Clone()
	sensitiveHeaders := []string{"Authorization", "Cookie", "Set-Cookie", "Proxy-Authorization", "WWW-Authenticate"}
	for _, key := range sensitiveHeaders {
		if header.Get(key) != "" {
			header.Set(key, "***")
		}
	}

	t.logger.Trace("roundtrip",
		"url", req.URL,
		"method", req.Method,
		"header", header,
	)
	return t.RoundTripper.RoundTrip(req)
}
