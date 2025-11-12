package maxworkersattr

import (
	"fmt"
	"os"
	"strconv"

	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	// TransferWorkersEnvVar is the environment variable to configure the number of transfer workers.
	TransferWorkersEnvVar = "OCM_TRANSFER_WORKER_COUNT"
	// ATTR_KEY is the full unique key for the max workers attribute.
	// This key should reflect its location or purpose within the ocm.software domain.
	ATTR_KEY = "ocm.software/ocm/api/ocm/extensions/attrs/maxworkers"
	// ATTR_SHORT is a shorter alias for the max workers attribute, useful for CLI.
	ATTR_SHORT = "maxworkers"

	// SingleWorker is 1. If the user doesn't specify the attribute,
	// Get() will return this value, signaling the calling code to use only one worker at a time.
	// Additionally, this can be used to signal that coding should run sequentially, so that ordering is guaranteed.
	SingleWorker uint = 1
)

func init() {
	// This function runs automatically when the package is imported.
	// It registers your attribute type with the OCM data context.
	if err := datacontext.RegisterAttributeType(ATTR_KEY, AttributeType{}, ATTR_SHORT); err != nil {
		panic(err)
	}
}

// AttributeType implements the datacontext.AttributeType interface for max workers.
type AttributeType struct{}

// Name returns the full unique key of the attribute.
func (a AttributeType) Name() string {
	return ATTR_KEY
}

// Description provides documentation for the attribute, visible in help messages.
func (a AttributeType) Description() string {
	return `
*integer*
Specifies the maximum number of concurrent workers to use for resource and source
transfer operations. This can influence performance and resource consumption.
A value of 0 (or not specified) indicates auto-detection based on CPU cores.
WARNING: This is an experimental feature and may cause unexpected issues.
`
}

// Encode converts the attribute's Go value (uint) to its marshaled byte representation (e.g., JSON).
func (a AttributeType) Encode(v interface{}, marshaller runtime.Marshaler) ([]byte, error) {
	val, ok := v.(uint) // Expecting a uint for number of workers
	if !ok {
		// Attempt to convert from int if it's passed as int (common Go numeric type)
		if intVal, ok := v.(int); ok {
			if intVal < 0 {
				return nil, fmt.Errorf("negative integer for maxworkers not allowed")
			}
			val = uint(intVal)
		} else {
			return nil, fmt.Errorf("unsigned integer (uint) or integer (int) required for maxworkers")
		}
	}
	return marshaller.Marshal(val) // Marshal the uint value
}

// Decode converts the marshaled byte representation (e.g., JSON) to the attribute's Go value (uint).
func (a AttributeType) Decode(data []byte, unmarshaller runtime.Unmarshaler) (interface{}, error) {
	var value uint // Decode into a uint
	err := unmarshaller.Unmarshal(data, &value)
	if err != nil {
		var s string
		if e := unmarshaller.Unmarshal(data, &s); e == nil {
			parsedVal, err := strconv.ParseUint(s, 10, 32)
			if err == nil {
				return uint(parsedVal), nil
			}
		}
		return nil, fmt.Errorf("failed to decode maxworkers as uint: %w", err)
	}
	return value, nil
}

////////////////////////////////////////////////////////////////////////////////

func Get(ctx datacontext.Context) (uint, error) {
	val, err := get(ctx)
	if err != nil {
		return 0, err
	}
	if val > SingleWorker {
		ctx.Logger().Warn(ATTR_SHORT + " attribute is not set to a single worker, this is experimental and may cause unexpected issues")
	}
	return val, nil
}

func get(ctx datacontext.Context) (uint, error) {
	a := ctx.GetAttributes().GetAttribute(ATTR_KEY)
	if a != nil {
		if val, ok := a.(uint); ok {
			return val, nil // Return the user-specified value (can be 0 if user explicitly set it to 0).
		} else {
			return 0, fmt.Errorf("unexpected type %T for maxworkers attribute, expected uint", a)
		}
	}
	if val, foundInEnv := os.LookupEnv(TransferWorkersEnvVar); foundInEnv {
		parsedFromEnv, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			return 0, fmt.Errorf("failed to parse %s environment variable: %w", TransferWorkersEnvVar, err)
		}
		return uint(parsedFromEnv), nil
	}
	return SingleWorker, nil
}

func Set(ctx datacontext.Context, workers uint) error {
	return ctx.GetAttributes().SetAttribute(ATTR_KEY, workers)
}
