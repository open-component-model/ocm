package featuregatesattr

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/mandelsoft/goutils/general"
	"sigs.k8s.io/yaml"

	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	ATTR_SHORT = "featuregates"
	ATTR_KEY   = "ocm.software/ocm/" + ATTR_SHORT
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
*featuregates* Enable/Disable optional features of the OCM library.
Optionally, particular features modes and attributes can be configured, if
supported by the feature implementation.
`
}

func (a AttributeType) Encode(v interface{}, marshaller runtime.Marshaler) ([]byte, error) {
	switch t := v.(type) {
	case *Attribute:
		return json.Marshal(v)
	case string:
		_, err := a.Decode([]byte(t), runtime.DefaultYAMLEncoding)
		if err != nil {
			return nil, err
		}
		return []byte(t), nil
	case []byte:
		_, err := a.Decode(t, runtime.DefaultYAMLEncoding)
		if err != nil {
			return nil, err
		}
		return t, nil
	default:
		return nil, fmt.Errorf("feature gate config required")
	}
}

func (a AttributeType) Decode(data []byte, unmarshaller runtime.Unmarshaler) (interface{}, error) {
	var c Attribute
	err := yaml.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

////////////////////////////////////////////////////////////////////////////////

const FEATURE_DISABLED = "off"

type Attribute struct {
	lock sync.Mutex

	Features map[string]*FeatureGate `json:"features"`
}

// FeatureGate store settings for a particular feature gate.
// To be extended by additional config possibility.
// Default behavior is to be enabled if entry is given
// for a feature name and mode is not equal *off*.
type FeatureGate struct {
	Mode       string                     `json:"mode"`
	Attributes map[string]json.RawMessage `json:"attributes,omitempty"`
}

func New() *Attribute {
	return &Attribute{Features: map[string]*FeatureGate{}}
}

func (a *Attribute) EnableFeature(name string, state *FeatureGate) {
	a.lock.Lock()
	defer a.lock.Unlock()

	if state == nil {
		state = &FeatureGate{}
	}
	if state.Mode == FEATURE_DISABLED {
		state.Mode = ""
	}
	a.Features[name] = state
}

func (a *Attribute) SetFeature(name string, state *FeatureGate) {
	a.lock.Lock()
	defer a.lock.Unlock()

	if state == nil {
		state = &FeatureGate{}
	}
	a.Features[name] = state
}

func (a *Attribute) DisableFeature(name string) {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.Features[name] = &FeatureGate{Mode: "off"}
}

func (a *Attribute) DefaultFeature(name string) {
	a.lock.Lock()
	defer a.lock.Unlock()

	delete(a.Features, name)
}

func (a *Attribute) IsEnabled(name string, def ...bool) bool {
	return a.GetFeature(name, def...).Mode != FEATURE_DISABLED
}

func (a *Attribute) GetFeature(name string, def ...bool) *FeatureGate {
	a.lock.Lock()
	defer a.lock.Unlock()

	g, ok := a.Features[name]
	if !ok {
		g = &FeatureGate{}
		if !general.Optional(def...) {
			g.Mode = FEATURE_DISABLED
		}
	}
	return g
}

////////////////////////////////////////////////////////////////////////////////

func Get(ctx datacontext.Context) *Attribute {
	v := ctx.GetAttributes().GetAttribute(ATTR_KEY)
	if v == nil {
		v = New()
	}
	return v.(*Attribute)
}

func Set(ctx datacontext.Context, c *Attribute) {
	ctx.GetAttributes().SetAttribute(ATTR_KEY, c)
}

var lock sync.Mutex

func get(ctx datacontext.Context) *Attribute {
	attrs := ctx.GetAttributes()
	v := attrs.GetAttribute(ATTR_KEY)

	if v == nil {
		v = New()
		attrs.SetAttribute(ATTR_KEY, v)
	}
	return v.(*Attribute)
}

func SetFeature(ctx datacontext.Context, name string, state *FeatureGate) {
	lock.Lock()
	defer lock.Unlock()

	get(ctx).SetFeature(name, state)
}

func EnableFeature(ctx datacontext.Context, name string, state *FeatureGate) {
	lock.Lock()
	defer lock.Unlock()

	get(ctx).EnableFeature(name, state)
}

func DisableFeature(ctx datacontext.Context, name string) {
	lock.Lock()
	defer lock.Unlock()

	get(ctx).DisableFeature(name)
}

func DefaultFeature(ctx datacontext.Context, name string) {
	lock.Lock()
	defer lock.Unlock()

	get(ctx).DefaultFeature(name)
}
