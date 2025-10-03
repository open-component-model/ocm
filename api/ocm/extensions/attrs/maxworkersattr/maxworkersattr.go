package maxworkersattr

import (
	"fmt"
	"strconv"

	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/utils/runtime" 
)

const (
	// ATTR_KEY is the full unique key for the max workers attribute.
	// This key should reflect its location or purpose within the ocm.software domain.
	ATTR_KEY = "ocm.software/ocm/api/ocm/extensions/attrs/maxworkers"
	// ATTR_SHORT is a shorter alias for the max workers attribute, useful for CLI.
	ATTR_SHORT = "maxworkers"

	// InternalDefault is now 0. If the user doesn't specify the attribute,
	// Get() will return 0, signaling the calling code to use CPU-based auto-detection.
	InternalDefault uint = 0
)

func init() {
	// This function runs automatically when the package is imported.
	// It registers your attribute type with the OCM data context.
	datacontext.RegisterAttributeType(ATTR_KEY, AttributeType{}, ATTR_SHORT)
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

func Get(ctx datacontext.Context) uint {
	a := ctx.GetAttributes().GetAttribute(ATTR_KEY)
	if a == nil {
		// If the attribute is NOT explicitly set by the user, return 0.
		// This 0 will signal the calling code to use the CPU-based auto-detection.
		return 0
	}
	if val, ok := a.(uint); ok {
		return val // Return the user-specified value (can be 0 if user explicitly set it to 0).
	}
	// Fallback in case of type mismatch (should ideally not happen with correct Encode/Decode)
	return 0 // Default to 0 for auto-detection in case of error
}

func Set(ctx datacontext.Context, workers uint) error {
	return ctx.GetAttributes().SetAttribute(ATTR_KEY, workers)
}

