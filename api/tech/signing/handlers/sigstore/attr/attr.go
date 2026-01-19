package attr

import (
	"errors"
	"fmt"

	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	ATTR_KEY   = "ocm.software/signing/sigstore"
	ATTR_SHORT = "sigstore"
)

var defaultAttr = Attribute{
	// https://github.com/tektoncd/chains/blob/main/docs/config.md#keyless-signing-with-fulcio
	FulcioURL:    "https://fulcio.sigstore.dev",
	RekorURL:     "https://rekor.sigstore.dev",
	OIDCIssuer:   "https://oauth2.sigstore.dev/auth",
	OIDCClientID: "sigstore",
}

func init() {
	datacontext.RegisterAttributeType(ATTR_KEY, AttributeType{}, ATTR_SHORT)
}

// AttributeType represents the attribute functionality.
type AttributeType struct{}

// Attribute represents the attribute data.
type Attribute struct {
	FulcioURL    string `json:"fulcioURL"`
	RekorURL     string `json:"rekorURL"`
	OIDCIssuer   string `json:"OIDCIssuer"`
	OIDCClientID string `json:"OIDCClientID"`
}

// Name returns the attribute name.
func (a AttributeType) Name() string {
	return ATTR_KEY
}

// Description returns a description of the attribute.
func (a AttributeType) Description() string {
	return `
*sigstore config* Configuration to use for sigstore based signing.

This configuration applies to both <code>sigstore</code> (legacy) and <code>sigstore-v2</code> signing algorithms.
The difference between the algorithms is transparent to configuration but affects how signatures are stored in Rekor:
- <code>sigstore</code>: stores only the public key in the Rekor entry (does not comply with Sigstore Bundle specification).
- <code>sigstore-v2</code>: stores the Fulcio certificate in the Rekor entry (complies with Sigstore Bundle specification).

The following fields are used.
- *<code>fulcioURL</code>* *string*  default is https://fulcio.sigstore.dev
- *<code>rekorURL</code>* *string*  default is https://rekor.sigstore.dev
- *<code>OIDCIssuer</code>* *string*  default is https://oauth2.sigstore.dev/auth
- *<code>OIDCClientID</code>* *string*  default is sigstore
`
}

// Encode marshals an attribute.
func (AttributeType) Encode(v interface{}, marshaler runtime.Marshaler) ([]byte, error) {
	if marshaler == nil {
		marshaler = runtime.DefaultJSONEncoding
	}

	result, ok := v.(*Attribute)
	if !ok {
		return nil, errors.New("sigstore attribute required")
	}

	return marshaler.Marshal(result)
}

// Decode unmarshals an attribute.
func (a AttributeType) Decode(data []byte, unmarshaler runtime.Unmarshaler) (interface{}, error) {
	if unmarshaler == nil {
		unmarshaler = runtime.DefaultJSONEncoding
	}

	attr := &Attribute{}
	err := unmarshaler.Unmarshal(data, attr)
	if err != nil {
		return nil, fmt.Errorf("invalid attribute value for %s: %w", ATTR_KEY, err)
	}

	return attr, nil
}

// Get returns the attributes.
func Get(ctx datacontext.Context) *Attribute {
	v := ctx.GetAttributes().GetAttribute(ATTR_KEY)
	if v == nil {
		return &defaultAttr
	}
	a, ok := v.(*Attribute)
	if !ok {
		return &defaultAttr
	}
	return a
}

// Set sets the attributes.
func Set(ctx datacontext.Context, a *Attribute) error {
	attrs := ctx.GetAttributes()
	return attrs.SetAttribute(ATTR_KEY, a)
}
