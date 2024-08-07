package internal

import (
	"context"
	"reflect"
	"strings"

	"ocm.software/ocm/api/config"
	cfgcpi "ocm.software/ocm/api/config/cpi"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/runtime"
)

const CONTEXT_TYPE = "oci" + datacontext.OCM_CONTEXT_SUFFIX

const CommonTransportFormat = "CommonTransportFormat"

type ContextProvider interface {
	OCIContext() Context
}

type Context interface {
	datacontext.Context
	config.ContextProvider
	credentials.ContextProvider
	ContextProvider

	RepositorySpecHandlers() RepositorySpecHandlers
	MapUniformRepositorySpec(u *UniformRepositorySpec) (RepositorySpec, error)

	RepositoryTypes() RepositoryTypeScheme

	RepositoryForSpec(spec RepositorySpec, creds ...credentials.CredentialsSource) (Repository, error)
	RepositoryForConfig(data []byte, unmarshaler runtime.Unmarshaler, creds ...credentials.CredentialsSource) (Repository, error)
	RepositorySpecForConfig(data []byte, unmarshaler runtime.Unmarshaler) (RepositorySpec, error)

	GetAlias(name string) RepositorySpec
	SetAlias(name string, spec RepositorySpec)
}

var key = reflect.TypeOf(_context{})

// DefaultContext is the default context initialized by init functions.
var DefaultContext = Builder{}.New(datacontext.MODE_SHARED)

// ForContext returns the Context to use for context.Context.
// This is either an explicit context or the default context.
func ForContext(ctx context.Context) Context {
	c, _ := datacontext.ForContextByKey(ctx, key, DefaultContext)
	return c.(Context)
}

func FromProvider(p ContextProvider) Context {
	if p == nil {
		return nil
	}
	return p.OCIContext()
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

type _context struct {
	_InternalContext
	updater cfgcpi.Updater

	credentials credentials.Context

	knownRepositoryTypes RepositoryTypeScheme
	specHandlers         RepositorySpecHandlers
	aliases              map[string]RepositorySpec
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

func newContext(credctx credentials.Context, reposcheme RepositoryTypeScheme, specHandlers RepositorySpecHandlers, delegates datacontext.Delegates) Context {
	c := &_context{
		credentials:          datacontext.PersistentContextRef(credctx),
		knownRepositoryTypes: reposcheme,
		specHandlers:         specHandlers,
		aliases:              map[string]RepositorySpec{},
	}
	c._InternalContext = datacontext.NewContextBase(c, CONTEXT_TYPE, key, credctx.ConfigContext().GetAttributes(), delegates)
	c.updater = cfgcpi.NewUpdaterForFactory(credctx.ConfigContext(), c.OCIContext)
	return newView(c, true)
}

func (c *_context) CreateView() Context {
	return newView(c, true)
}

func (c *_context) OCIContext() Context {
	return newView(c)
}

func (c *_context) Update() error {
	return c.updater.Update()
}

func (c *_context) AttributesContext() datacontext.AttributesContext {
	return c.credentials.AttributesContext()
}

func (c *_context) ConfigContext() config.Context {
	return c.updater.GetContext()
}

func (c *_context) CredentialsContext() credentials.Context {
	return c.credentials
}

func (c *_context) RepositoryTypes() RepositoryTypeScheme {
	return c.knownRepositoryTypes
}

func (c *_context) RepositorySpecHandlers() RepositorySpecHandlers {
	return c.specHandlers
}

func (c *_context) MapUniformRepositorySpec(u *UniformRepositorySpec) (RepositorySpec, error) {
	return c.specHandlers.MapUniformRepositorySpec(c.OCIContext(), u)
}

func (c *_context) RepositorySpecForConfig(data []byte, unmarshaler runtime.Unmarshaler) (RepositorySpec, error) {
	return c.knownRepositoryTypes.Decode(data, unmarshaler)
}

func (c *_context) RepositoryForSpec(spec RepositorySpec, creds ...credentials.CredentialsSource) (Repository, error) {
	cred, err := credentials.CredentialsChain(creds).Credentials(c.CredentialsContext())
	if err != nil {
		return nil, err
	}
	return spec.Repository(c.OCIContext(), cred)
}

func (c *_context) RepositoryForConfig(data []byte, unmarshaler runtime.Unmarshaler, creds ...credentials.CredentialsSource) (Repository, error) {
	spec, err := c.knownRepositoryTypes.Decode(data, unmarshaler)
	if err != nil {
		return nil, err
	}
	return c.RepositoryForSpec(spec, creds...)
}

func (c *_context) GetAlias(name string) RepositorySpec {
	err := c.updater.Update()
	if err != nil {
		return nil
	}
	c.updater.RLock()
	defer c.updater.RUnlock()
	spec := c.aliases[name]
	if spec == nil && strings.HasSuffix(name, ".alias") {
		spec = c.aliases[name[:len(name)-6]]
	}
	return spec
}

func (c *_context) SetAlias(name string, spec RepositorySpec) {
	c.updater.Lock()
	defer c.updater.Unlock()
	c.aliases[name] = spec
}
