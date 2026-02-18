package config

import (
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ocm.software/ocm/api/datacontext/attrs/httptimeoutattr"
	"ocm.software/ocm/api/ocm"
)

func TestTimeoutFlag(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected time.Duration
	}{
		{
			name:     "not set defaults to 30s",
			args:     []string{},
			expected: httptimeoutattr.DefaultTimeout,
		},
		{
			name:     "set to 30s",
			args:     []string{"--timeout", "30s"},
			expected: 30 * time.Second,
		},
		{
			name:     "set to 5m",
			args:     []string{"--timeout", "5m"},
			expected: 5 * time.Minute,
		},
		{
			name:     "set to 1h",
			args:     []string{"--timeout", "1h"},
			expected: 1 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := New()
			fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
			cfg.AddFlags(fs)

			err := fs.Parse(tt.args)
			require.NoError(t, err)

			ctx := ocm.New()
			_, err = cfg.Evaluate(ctx, true)
			require.NoError(t, err)

			assert.Equal(t, tt.expected, httptimeoutattr.Get(ctx))
		})
	}
}

func TestTimeoutFlag_InvalidValue(t *testing.T) {
	cfg := New()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	cfg.AddFlags(fs)

	err := fs.Parse([]string{"--timeout", "notaduration"})
	require.NoError(t, err)

	ctx := ocm.New()
	_, err = cfg.Evaluate(ctx, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid timeout value \"notaduration\": use a duration string like 30s, 5m, or 1h")
}
