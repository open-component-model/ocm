package maxworkersattr

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"

	"ocm.software/ocm/api/datacontext"
	rtruntime "ocm.software/ocm/api/utils/runtime"
)

const (
	// TransferWorkersEnvVar defines the environment variable that configures
	// the maximum number of concurrent transfer workers.
	//
	// If set to a positive integer value, that number of concurrent workers is used.
	// If set to "auto", the number of logical CPU cores on the system is used.
	// If unset, the attribute value (if any) or the default of SingleWorker (1) is used.
	TransferWorkersEnvVar = "OCM_TRANSFER_WORKER_COUNT"

	// ATTR_KEY is the globally unique key under which this attribute is registered
	// in the OCM data context. It follows the ocm.software naming convention.
	ATTR_KEY = "ocm.software/ocm/api/ocm/extensions/attrs/maxworkers"

	// ATTR_SHORT is the short alias of the attribute, suitable for CLI or YAML use.
	ATTR_SHORT = "maxworkers"

	// SingleWorker is the default number of workers (1) used when no configuration
	// is provided. This mode guarantees deterministic ordering of operations.
	SingleWorker uint = 1

	// autoLiteral is the string literal used to indicate that the number of workers
	// should be automatically determined based on the number of logical CPU cores.
	autoLiteral = "auto"
)

func init() {
	if err := datacontext.RegisterAttributeType(ATTR_KEY, AttributeType{}, ATTR_SHORT); err != nil {
		panic(err)
	}
}

// AttributeType implements the datacontext.AttributeType interface for
// the `maxworkers` attribute. It controls the maximum concurrency used
// during resource and source transfer operations.
type AttributeType struct{}

// Name returns the globally unique key for this attribute.
func (a AttributeType) Name() string { return ATTR_KEY }

// Description provides extended docs for this attribute.
func (a AttributeType) Description() string {
	return `
*integer* or *"auto"*
Specifies the maximum number of concurrent workers to use for resource and source,
as well as reference transfer operations.

Supported values:
  - A positive integer: use exactly that number of workers.
  - The string "auto": automatically use the number of logical CPU cores.
  - Zero or omitted: fall back to single-worker mode (1). This is the default.
    This mode guarantees deterministic ordering of operations.

Precedence:
  1. Attribute set in the current OCM context.
  2. Environment variable OCM_TRANSFER_WORKER_COUNT.
  3. Default value (1).

WARNING: This is an experimental feature and may cause unexpected behavior
depending on workload concurrency. Values above 1 may result in non-deterministic
transfer ordering.
`
}

// Encode converts the attribute's Go value into its marshaled representation.
// It supports uint, int, and string ("auto") forms.
func (a AttributeType) Encode(v interface{}, m rtruntime.Marshaler) ([]byte, error) {
	switch val := v.(type) {
	case uint:
		return m.Marshal(val)
	case int:
		if val < 0 {
			return nil, fmt.Errorf("negative integer for %s not allowed", ATTR_SHORT)
		}
		return m.Marshal(uint(val))
	case string:
		if val != autoLiteral {
			return nil, fmt.Errorf("invalid string value for %s: %q", ATTR_SHORT, val)
		}
		return m.Marshal(val)
	default:
		return nil, fmt.Errorf("unsupported type %T for %s", v, ATTR_SHORT)
	}
}

// Decode converts marshaled bytes back into Go form (either uint or "auto").
func (a AttributeType) Decode(data []byte, unmarshaller rtruntime.Unmarshaler) (interface{}, error) {
	// Try uint first (e.g., `6`)
	var value uint
	if err := unmarshaller.Unmarshal(data, &value); err == nil {
		return value, nil
	}

	// Try string next (e.g., `"auto"` or `"6"`)
	var s string
	if err := unmarshaller.Unmarshal(data, &s); err == nil {
		switch s {
		case autoLiteral:
			return s, nil
		default:
			if parsedVal, err := strconv.ParseUint(s, 10, 32); err == nil {
				return uint(parsedVal), nil
			}
			return nil, fmt.Errorf("invalid string value for %s: %q", ATTR_SHORT, s)
		}
	}

	return nil, fmt.Errorf("failed to decode %s", ATTR_SHORT)
}

////////////////////////////////////////////////////////////////////////////////

// Get returns the resolved number of concurrent transfer workers from the context.
// Resolution order:
//  1. Attribute value (ctx)
//  2. Environment variable OCM_TRANSFER_WORKER_COUNT
//  3. Default SingleWorker (1)
//
// The resolver only auto-detects CPUs if the value is exactly "auto".
// Any 0 resolves to SingleWorker.
func Get(ctx datacontext.Context) (uint, error) {
	var val = SingleWorker
	var err error

	if attribute := ctx.GetAttributes().GetAttribute(ATTR_KEY); attribute != nil {
		val, err = resolveWorkers(attribute)
	} else if env, ok := os.LookupEnv(TransferWorkersEnvVar); ok {
		val, err = resolveWorkers(env)
	}

	if err != nil {
		return 0, err
	}

	if val > SingleWorker {
		warnUnstableOnce.Do(func() {
			ctx.Logger().Warn(ATTR_SHORT + " attribute is set to more than 1 worker, this may cause unexpected behavior")
		})
	}

	// 3) Default
	return val, nil
}

// warnUnstableOnce ensures we only log only one warning if the attribute is retrieved multiple times.
var warnUnstableOnce sync.Once

// Set stores the attribute after validation via the unified resolver.
// Accepts uint, int>=0, or the string "auto".
func Set(ctx datacontext.Context, workers any) error {
	val, err := resolveWorkers(workers)
	if err != nil {
		return err
	}
	return ctx.GetAttributes().SetAttribute(ATTR_KEY, val)
}

////////////////////////////////////////////////////////////////////////////////

// resolveWorkers normalizes all supported input forms into a concrete uint.
// Supported forms:
//   - uint, int >= 0
//   - string "auto" → runtime.NumCPU()
//   - numeric string (e.g. "4") → parsed value
//   - 0 → SingleWorker
func resolveWorkers(v any) (uint, error) {
	switch t := v.(type) {
	case nil:
		return SingleWorker, nil

	case uint:
		if t == 0 {
			return SingleWorker, nil
		}
		return t, nil

	case int:
		if t < 0 {
			return 0, fmt.Errorf("%s cannot be negative", ATTR_SHORT)
		}
		if t == 0 {
			return SingleWorker, nil
		}
		return uint(t), nil

	case string:
		if t == autoLiteral {
			n := runtime.NumCPU()
			if n <= 0 {
				return SingleWorker, nil
			}
			return uint(n), nil
		}
		// Try numeric string conversion
		if parsed, err := strconv.ParseUint(t, 10, 32); err == nil {
			if parsed == 0 {
				return SingleWorker, nil
			}
			return uint(parsed), nil
		}
		return 0, fmt.Errorf("invalid string value for %s: %q", ATTR_SHORT, t)

	default:
		return 0, fmt.Errorf("unexpected %s type %T", ATTR_SHORT, v)
	}
}
