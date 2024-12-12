package regclient

import (
	"testing"

	"github.com/regclient/regclient/config"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	host := config.HostNew()
	n := New(ClientOptions{Host: []config.Host{
		*host,
	}})
	require.NotNil(t, n)
}
