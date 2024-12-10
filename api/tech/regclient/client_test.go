package regclient

import (
	"testing"

	"github.com/regclient/regclient/config"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	n := New(ClientOptions{Host: config.HostNew()})
	require.NotNil(t, n)
}
