package ocireg

import (
	"crypto/tls"
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHTTPTransportClonesDefaultTransport(t *testing.T) {
	t.Parallel()

	base, ok := http.DefaultTransport.(*http.Transport)
	require.True(t, ok, "default transport must be *http.Transport")

	conf := &tls.Config{MinVersion: tls.VersionTLS12}

	got := newHTTPTransport(conf)
	require.NotNil(t, got)

	assert.NotSame(t, base, got)
	assert.Same(t, conf, got.TLSClientConfig)
	assert.Equal(t, base.ForceAttemptHTTP2, got.ForceAttemptHTTP2)
	assert.Equal(t, base.MaxIdleConns, got.MaxIdleConns)
	assert.Equal(t, base.MaxIdleConnsPerHost, got.MaxIdleConnsPerHost)
	assert.Equal(t, base.MaxConnsPerHost, got.MaxConnsPerHost)
	assert.Equal(t, base.IdleConnTimeout, got.IdleConnTimeout)
	assert.Equal(t, base.TLSHandshakeTimeout, got.TLSHandshakeTimeout)
	assert.Equal(t, base.ExpectContinueTimeout, got.ExpectContinueTimeout)

	if base.Proxy != nil {
		require.NotNil(t, got.Proxy)
		assert.Equal(t, reflect.ValueOf(base.Proxy).Pointer(), reflect.ValueOf(got.Proxy).Pointer())
	}
}
