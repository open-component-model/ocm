package compatattr

import (
	"fmt"

	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	ATTR_KEY   = "github.com/mandelsoft/ocm/compat"
	ATTR_SHORT = "compat"
)

func init() {
	datacontext.RegisterAttributeType(ATTR_KEY, AttributeType{}, ATTR_SHORT)
}

type AttributeType struct{}

func (a AttributeType) Name() string {
	return ATTR_KEY
}

func (a AttributeType) Description() string {
	return `
*bool*
Compatibility mode: Avoid generic local access methods and prefer type specific ones.
`
}

func (a AttributeType) Encode(v interface{}, marshaller runtime.Marshaler) ([]byte, error) {
	if _, ok := v.(bool); !ok {
		return nil, fmt.Errorf("boolean required")
	}
	return marshaller.Marshal(v)
}

func (a AttributeType) Decode(data []byte, unmarshaller runtime.Unmarshaler) (interface{}, error) {
	var value bool
	err := unmarshaller.Unmarshal(data, &value)
	return value, err
}

////////////////////////////////////////////////////////////////////////////////

func Get(ctx datacontext.Context) bool {
	a := ctx.GetAttributes().GetAttribute(ATTR_KEY)
	if a == nil {
		return false
	}
	return a.(bool)
}

func Set(ctx datacontext.Context, flag bool) error {
	return ctx.GetAttributes().SetAttribute(ATTR_KEY, flag)
}
