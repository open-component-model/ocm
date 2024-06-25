package clicfgattr

import (
	"encoding/json"
	"fmt"

	"sigs.k8s.io/yaml"

	"github.com/open-component-model/ocm/pkg/contexts/config"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	ATTR_KEY   = "ocm.software/cliconfig"
	ATTR_SHORT = "cliconfig"
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
*cliconfigr* Configuration Object passed to command line pluging.
`
}

func (a AttributeType) Encode(v interface{}, marshaller runtime.Marshaler) ([]byte, error) {
	switch c := v.(type) {
	case config.Config:
		return json.Marshal(v)
	case []byte:
		if _, err := a.Decode(c, nil); err != nil {
			return nil, err
		}
		return c, nil
	default:
		return nil, fmt.Errorf("config object required")
	}
}

func (a AttributeType) Decode(data []byte, _ runtime.Unmarshaler) (interface{}, error) {
	var c config.GenericConfig
	err := yaml.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

////////////////////////////////////////////////////////////////////////////////

func Get(ctx datacontext.Context) config.Config {
	v := ctx.GetAttributes().GetAttribute(ATTR_KEY)
	if v == nil {
		return nil
	}
	return v.(config.Config)
}

func Set(ctx datacontext.Context, c config.Config) {
	ctx.GetAttributes().SetAttribute(ATTR_KEY, c)
}
