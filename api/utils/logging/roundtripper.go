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
	t.logger.Trace("roundtrip",
		"url", req.URL,
		"method", req.Method,
	)
	return t.RoundTripper.RoundTrip(req)
}
