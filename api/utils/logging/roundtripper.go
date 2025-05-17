package logging

import (
	"net/http"
	"net/url"

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
	reqURL := *req.URL
	if _, set := reqURL.User.Password(); set {
		reqURL.User = url.UserPassword(reqURL.User.Username(), "****")
	}
	t.logger.Trace("roundtrip",
		"url", reqURL.String(),
		"method", req.Method,
	)
	return t.RoundTripper.RoundTrip(req)
}
