package internal

import (
	"context"
	"reflect"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/utils"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/runtime"
)

// OCM_CONFIG_TYPE_SUFFIX is the standard suffix used for configuration
// types provided by this library.
const OCM_CONFIG_TYPE_SUFFIX = ".config" + common.OCM_TYPE_GROUP_SUFFIX

type ConfigSelector interface {
	Select(Config) bool
}
type ConfigSelectorFunction func(Config) bool

func (f ConfigSelectorFunction) Select(cfg Config) bool { return f(cfg) }

var AllConfigs = AppliedConfigSelectorFunction(func(*AppliedConfig) bool { return true })

const AllGenerations int64 = 0

const CONTEXT_TYPE = "config" + datacontext.OCM_CONTEXT_SUFFIX

type ContextProvider interface {
	ConfigContext() Context
}

type Context interface {
	datacontext.Context
	ContextProvider

	AttributesContext() datacontext.AttributesContext

	// Info provides the context for nested configuration evaluation
	Info() string
	// WithInfo provides the same context with additional nesting info
	WithInfo(desc string) Context

	ConfigTypes() ConfigTypeScheme

	// SkipUnknownConfig can be used to control the behaviour
	// for processing unknown configuration object types.
	// It returns the previous mode valid before setting the
	// new one.
	SkipUnknownConfig(bool) bool

	// Validate validates the applied configuration for not using
	// unknown configuration types, anymore. This can be used after setting
	// SkipUnknownConfig, to check whether there are still unknown types
	// which will be skipped. It does not provide information, whether
	// config objects were skipped for previous object configuration
	// requests.
	Validate() error

	// GetConfigForData deserialize configuration objects for known
	// configuration types.
	GetConfigForData(data []byte, unmarshaler runtime.Unmarshaler) (Config, error)

	// ApplyData applies the config given by a byte stream to the config store
	// If the config type is not known, a generic config is stored and returned.
	// In this case an unknown error for kind KIND_CONFIGTYPE is returned.
	ApplyData(data []byte, unmarshaler runtime.Unmarshaler, desc string) (Config, error)
	// ApplyConfig applies the config to the config store
	ApplyConfig(spec Config, desc string) error

	GetConfigForType(generation int64, typ string) (int64, []Config)
	GetConfigForName(generation int64, name string) (int64, []Config)
	GetConfig(generation int64, selector ConfigSelector) (int64, []Config)

	AddConfigSet(name string, set *ConfigSet)
	ApplyConfigSet(name string) error

	// Reset all configs applied so far, subsequent calls to ApplyTo will
	// only see configs applied after the last reset.
	Reset() int64
	// Generation return the actual config generation.
	// this is a strictly increasing number, regardless of the number
	// of Reset calls.
	Generation() int64
	// ApplyTo applies all configurations applied after the last reset with
	// a generation larger than the given watermark to the specified target.
	// A target may be any object. The applied configuration objects decide
	// on their own whether they are applicable for the given target.
	// The generation of the last applied object is returned to be used as
	// new watermark.
	ApplyTo(gen int64, target interface{}) (int64, error)
}

var key = reflect.TypeOf(_context{})

// DefaultContext is the default context initialized by init functions.
var DefaultContext = Builder{}.New(datacontext.MODE_SHARED)

// FromContext returns the Context to use for context.Context.
// This is either an explicit context or the default context.
// The returned context incorporates the given context.
func FromContext(ctx context.Context) Context {
	c, _ := datacontext.ForContextByKey(ctx, key, DefaultContext)
	return c.(Context)
}

func FromProvider(p ContextProvider) Context {
	if p == nil {
		return nil
	}
	return p.ConfigContext()
}

func DefinedForContext(ctx context.Context) (Context, bool) {
	c, ok := datacontext.ForContextByKey(ctx, key, DefaultContext)
	if c != nil {
		return c.(Context), ok
	}
	return nil, ok
}

////////////////////////////////////////////////////////////////////////////////

type _InternalContext = datacontext.InternalContext

type coreContext struct {
	_InternalContext
	updater Updater

	sharedAttributes datacontext.AttributesContext

	knownConfigTypes ConfigTypeScheme

	configs           *ConfigStore
	skipUnknownConfig bool
}

type _context struct {
	*coreContext
	description string
}

var (
	_ Context                          = (*_context)(nil)
	_ datacontext.ViewCreator[Context] = (*_context)(nil)
)

// gcWrapper is used as garbage collectable
// wrapper for a context implementation
// to establish a runtime finalizer.
type gcWrapper struct {
	datacontext.GCWrapper
	*_context
}

func newView(c *_context, ref ...bool) Context {
	if utils.Optional(ref...) {
		return datacontext.FinalizedContext[gcWrapper](c)
	}
	return c
}

func (w *gcWrapper) SetContext(c *_context) {
	w._context = c
}

func newContext(shared datacontext.AttributesContext, reposcheme ConfigTypeScheme, delegates datacontext.Delegates) Context {
	c := &_context{
		coreContext: &coreContext{
			sharedAttributes: shared,
			knownConfigTypes: reposcheme,
			configs:          NewConfigStore(),
		},
	}
	c._InternalContext = datacontext.NewContextBase(c, CONTEXT_TYPE, key, shared.GetAttributes(), delegates)
	c.updater = NewUpdaterForFactory(c, c.ConfigContext) // provide target as new view to internal context
	datacontext.AssureUpdater(shared, NewUpdater(c, datacontext.PersistentContextRef(shared)))

	return newView(c, true)
}

func (c *_context) CreateView() Context {
	return newView(c, true)
}

func (c *_context) ConfigContext() Context {
	return newView(c)
}

func (c *_context) Update() error {
	return c.updater.Update()
}

var _ datacontext.Updater = (*_context)(nil)

func (c *_context) Info() string {
	return c.description
}

func (c *_context) WithInfo(desc string) Context {
	if c.description != "" {
		desc = desc + "--" + c.description
	}
	return newView(&_context{c.coreContext, desc})
}

func (c *_context) AttributesContext() datacontext.AttributesContext {
	c.updater.Update()
	return c.sharedAttributes
}

func (c *_context) ConfigTypes() ConfigTypeScheme {
	return c.knownConfigTypes
}

func (c *_context) SkipUnknownConfig(b bool) bool {
	old := c.skipUnknownConfig
	c.skipUnknownConfig = b
	return old
}

func (c *_context) ConfigForData(data []byte, unmarshaler runtime.Unmarshaler) (Config, error) {
	return c.knownConfigTypes.Decode(data, unmarshaler)
}

func (c *_context) GetConfigForData(data []byte, unmarshaler runtime.Unmarshaler) (Config, error) {
	spec, err := c.knownConfigTypes.Decode(data, unmarshaler)
	if err != nil {
		return nil, err
	}
	return spec, nil
}

func (c *_context) ApplyConfig(spec Config, desc string) error {
	var unknown error

	// use temporary view for outbound calls
	spec, err := (&AppliedConfig{config: spec}).eval(newView(c))
	if err != nil {
		if !errors.IsErrUnknownKind(err, KIND_CONFIGTYPE) {
			return errors.Wrapf(err, "%s", desc)
		}
		if !c.skipUnknownConfig {
			unknown = err
		}
		err = nil
	}

	c.configs.Apply(spec, desc)

	for {
		// apply directly and also indirectly described configurations
		if gen, in := c.updater.State(); err != nil || in || gen >= c.configs.Generation() {
			break
		}
		err = c.Update()
		if IsErrNoContext(err) {
			err = unknown
		}
	}

	return errors.Wrapf(err, "%s", desc)
}

func (c *_context) ApplyData(data []byte, unmarshaler runtime.Unmarshaler, desc string) (Config, error) {
	spec, err := c.knownConfigTypes.Decode(data, unmarshaler)
	if err != nil {
		return nil, err
	}
	return spec, c.ApplyConfig(spec, desc)
}

func (c *_context) selector(gen int64, selector ConfigSelector) AppliedConfigSelector {
	if gen <= 0 {
		return AppliedConfigSelectorFor(selector)
	}
	if selector == nil {
		return AppliedGenerationSelector(gen)
	}
	return AppliedAndSelector(AppliedGenerationSelector(gen), AppliedConfigSelectorFor(selector))
}

func (c *_context) Generation() int64 {
	return c.configs.Generation()
}

func (c *_context) Reset() int64 {
	return c.configs.Reset()
}

func (c *_context) ApplyTo(gen int64, target interface{}) (int64, error) {
	cur := c.configs.Generation()
	if cur <= gen {
		return gen, nil
	}
	cur, cfgs := c.configs.GetConfigForSelector(c, AppliedGenerationSelector(gen))

	list := errors.ErrListf("config apply errors")
	for _, cfg := range cfgs {
		err := cfg.config.ApplyTo(c.WithInfo(cfg.description), target)
		if c.skipUnknownConfig && errors.IsErrUnknownKind(err, KIND_CONFIGTYPE) {
			err = nil
		}
		err = errors.Wrapf(err, "%s", cfg.description)
		if !IsErrNoContext(err) {
			list.Add(err)
		}
	}
	return cur, list.Result()
}

func (c *_context) Validate() error {
	list := errors.ErrList()

	_, cfgs := c.configs.GetConfigForSelector(c, AllAppliedConfigs)
	for _, cfg := range cfgs {
		_, err := cfg.eval(newView(c))
		list.Add(err)
	}
	return list.Result()
}

func (c *_context) AddConfigSet(name string, set *ConfigSet) {
	c.configs.AddSet(name, set)
}

func (c *_context) ApplyConfigSet(name string) error {
	set := c.configs.GetSet(name)
	if set == nil {
		return errors.ErrUnknown(KIND_CONFIGSET, name)
	}
	desc := "config set " + name
	list := errors.ErrListf("applying %s", desc)
	for _, cfg := range set.Configurations {
		list.Add(c.ApplyConfig(cfg, desc))
	}
	return list.Result()
}

func (c *_context) GetConfig(gen int64, selector ConfigSelector) (int64, []Config) {
	gen, cfgs := c.configs.GetConfigForSelector(c, c.selector(gen, selector))
	return gen, cfgs.Configs()
}

func (c *_context) GetConfigForName(gen int64, name string) (int64, []Config) {
	gen, cfgs := c.configs.GetConfigForName(c, name, c.selector(gen, nil))
	return gen, cfgs.Configs()
}

func (c *_context) GetConfigForType(gen int64, typ string) (int64, []Config) {
	gen, cfgs := c.configs.GetConfigForType(c, typ, c.selector(gen, nil))
	return gen, cfgs.Configs()
}
