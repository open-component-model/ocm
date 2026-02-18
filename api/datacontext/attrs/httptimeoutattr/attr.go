package httptimeoutattr

import (
	"fmt"
	"time"

	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	ATTR_KEY   = "ocm.software/ocm/api/datacontext/attrs/httptimeout"
	ATTR_SHORT = "timeout"

	// DefaultTimeout is the default HTTP client timeout used when no
	// configuration is provided.
	DefaultTimeout = 30 * time.Second
)

func init() {
	datacontext.RegisterAttributeType(ATTR_KEY, AttributeType{}, ATTR_SHORT)
}

// AttributeType implements the datacontext.AttributeType interface for
// the httptimeout attribute.
type AttributeType struct{}

func (a AttributeType) Name() string {
	return ATTR_KEY
}

func (a AttributeType) Description() string {
	return `
*string*
Configures the timeout duration for HTTP client requests used to access
OCI registries and other remote endpoints. The value is specified as a
Go duration string (e.g. "30s", "5m", "1h").

If not set, the default timeout of 30s is used.
`
}

func (a AttributeType) Encode(v interface{}, marshaller runtime.Marshaler) ([]byte, error) {
	switch val := v.(type) {
	case time.Duration:
		return marshaller.Marshal(val.String())
	case string:
		if _, err := time.ParseDuration(val); err != nil {
			return nil, fmt.Errorf("invalid duration string for %s: %q", ATTR_SHORT, val)
		}
		return marshaller.Marshal(val)
	default:
		return nil, fmt.Errorf("duration or duration string required for %s, got %T", ATTR_SHORT, v)
	}
}

func (a AttributeType) Decode(data []byte, unmarshaller runtime.Unmarshaler) (interface{}, error) {
	var v interface{}
	if err := unmarshaller.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("failed to decode %s: %w", ATTR_SHORT, err)
	}

	switch value := v.(type) {
	case float64:
		return time.Duration(value), nil
	case string:
		d, err := time.ParseDuration(value)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout value %q for %s: must be a duration like 30s, 5m, or nanoseconds number: %w", value, ATTR_SHORT, err)
		}
		return d, nil
	default:
		return nil, fmt.Errorf("timeout for %s must be a duration string or nanoseconds number, got %T", ATTR_SHORT, v)
	}
}

////////////////////////////////////////////////////////////////////////////////

// Get returns the configured HTTP client timeout from the context.
// If not set, DefaultTimeout is returned.
func Get(ctx datacontext.Context) time.Duration {
	a := ctx.GetAttributes().GetAttribute(ATTR_KEY)
	if a == nil {
		return DefaultTimeout
	}
	return a.(time.Duration)
}

// Set stores the HTTP client timeout attribute in the context.
func Set(ctx datacontext.Context, d time.Duration) error {
	return ctx.GetAttributes().SetAttribute(ATTR_KEY, d)
}
