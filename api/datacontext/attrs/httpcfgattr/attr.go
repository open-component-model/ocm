package httpcfgattr

import (
	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/utils/runtime"
)

type (
	Context         = datacontext.AttributesContext
	ContextProvider = datacontext.ContextProvider
)

const (
	ATTR_KEY   = "ocm.software/ocm/api/datacontext/attrs/httptimeout"
	ATTR_SHORT = "httpcfg"
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
*JSON*
Configures HTTP client timeout settings for OCI registry and remote endpoint access.
Settings are provided as a JSON document matching the ` + ConfigType + ` config type.

For full control use the config file:
<pre>
    type: ` + ConfigType + `
    timeout: 0s
    tcpDialTimeout: 30s
    tcpKeepAlive: 30s
    tlsHandshakeTimeout: 10s
    idleConnTimeout: 90s
</pre>
`
}

func (a AttributeType) Encode(v interface{}, marshaller runtime.Marshaler) ([]byte, error) {
	attr, ok := v.(*Attribute)
	if !ok {
		return nil, errors.ErrInvalid("httpcfg attribute")
	}
	cfg := New()
	cfg.HTTPSettings = attr.settings

	return marshaller.Marshal(cfg)
}

func (a AttributeType) Decode(data []byte, unmarshaller runtime.Unmarshaler) (interface{}, error) {
	var value Config
	err := unmarshaller.Unmarshal(data, &value)
	if err != nil {
		return nil, err
	}

	attr := &Attribute{}
	err = value.ApplyToAttribute(attr)
	if err != nil {
		return nil, err
	}
	return attr, nil
}

////////////////////////////////////////////////////////////////////////////////

// Attribute holds the effective HTTP client settings for a context.
type Attribute struct {
	settings HTTPSettings
}

// GetHTTPSettings returns a pointer to the effective HTTP settings.
func (a *Attribute) GetHTTPSettings() *HTTPSettings {
	if a == nil {
		return &HTTPSettings{}
	}
	return &a.settings
}

////////////////////////////////////////////////////////////////////////////////

// Get returns the HTTP client attribute from the context.
// If not set, a default empty Attribute is created and stored.
func Get(ctx ContextProvider) *Attribute {
	return ctx.AttributesContext().GetAttributes().GetOrCreateAttribute(ATTR_KEY, func(datacontext.Context) interface{} {
		return &Attribute{}
	}).(*Attribute)
}

// Set stores the HTTP client attribute in the context.
func Set(ctx ContextProvider, attr *Attribute) error {
	return ctx.AttributesContext().GetAttributes().SetAttribute(ATTR_KEY, attr)
}
