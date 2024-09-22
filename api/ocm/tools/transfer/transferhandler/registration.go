package transferhandler

import (
	"sort"

	"github.com/mandelsoft/goutils/general"

	"ocm.software/ocm/api/datacontext"
	ocm "ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/utils/listformat"
	"ocm.software/ocm/api/utils/registrations"
)

const ATTR_TRANSFER_HANDLERS = "ocm.software/ocm/api/ocm/tools/transfer/transferhandlers"

func For(ctx ocm.ContextProvider) Registry {
	if ctx == nil {
		return DefaultRegistry
	}
	return ctx.OCMContext().GetAttributes().GetOrCreateAttribute(ATTR_TRANSFER_HANDLERS, create).(Registry)
}

func create(datacontext.Context) interface{} {
	return NewRegistry(DefaultRegistry)
}

func SetFor(ctx datacontext.Context, registry Registry) {
	ctx.GetAttributes().SetAttribute(ATTR_TRANSFER_HANDLERS, registry)
}

////////////////////////////////////////////////////////////////////////////////
// The creation handler is implemented by the RegistrationHandlerRegistry,
// but inśtead of executing a handler registration, like for the downloaders,
// the created handler is just returned,

type ByNameCreationHandler interface {
	ByName(ctx ocm.Context, path string, olist ...TransferOption) (bool, TransferHandler, error)
	GetHandlers(t *Target) registrations.HandlerInfos
}

type HandlerInfos = registrations.HandlerInfos

// byNameCreationHandler wraps the handler interface
// into the implementation type.
type byNameCreationHandler struct {
	handler ByNameCreationHandler
}

func (b byNameCreationHandler) GetHandlers(t *Target) HandlerInfos {
	return b.handler.GetHandlers(t)
}

func (b byNameCreationHandler) RegisterByName(path string, target *Target, _ registrations.HandlerConfig, opts ...TransferOption) (bool, error) {
	ok, h, err := b.handler.ByName(target.Context, path, opts...)
	target.Handler = h
	return ok, err
}

var _ registrations.HandlerRegistrationHandler[*Target, TransferOption] = (*byNameCreationHandler)(nil)

////////////////////////////////////////////////////////////////////////////////

type Target struct {
	Context ocm.Context
	Handler TransferHandler
}

func NewTarget(ctx ocm.ContextProvider) *Target {
	return &Target{Context: ctx.OCMContext()}
}

func RegisterHandlerRegistrationHandler(path string, handler ByNameCreationHandler) {
	DefaultRegistry.RegisterRegistrationHandler(path, handler)
}

func CreateByName(ctx ocm.ContextProvider, name string, opts ...TransferOption) (TransferHandler, error) {
	return For(ctx).ByName(ctx, name, opts...)
}

////////////////////////////////////////////////////////////////////////////////

var DefaultRegistry = NewRegistry()

type Registry interface {
	RegisterRegistrationHandler(path string, handler ByNameCreationHandler)

	ByName(ctx ocm.ContextProvider, name string, opts ...TransferOption) (TransferHandler, error)

	Copy() Registry
	AsHandlerRegistrationRegistry() registrations.HandlerRegistrationRegistry[*Target, TransferOption]
}

type _registry struct {
	registry registrations.HandlerRegistrationRegistry[*Target, TransferOption]
	base     Registry
}

func NewRegistry(base ...Registry) Registry {
	b := general.Optional(base...)
	return &_registry{registrations.NewHandlerRegistrationRegistry[*Target, TransferOption](asHandlerRegistrationRegistry(b)), b}
}

func (r *_registry) RegisterRegistrationHandler(path string, handler ByNameCreationHandler) {
	r.registry.RegisterRegistrationHandler(path, &byNameCreationHandler{handler})
}

func (r *_registry) ByName(ctx ocm.ContextProvider, name string, opts ...TransferOption) (TransferHandler, error) {
	target := NewTarget(ctx.OCMContext())
	_, err := r.registry.RegisterByName(name, target, nil, opts...)
	return target.Handler, err
}

func (r *_registry) Copy() Registry {
	return NewRegistry(r.base)
}

func (r *_registry) AsHandlerRegistrationRegistry() registrations.HandlerRegistrationRegistry[*Target, TransferOption] {
	return r.registry
}

func asHandlerRegistrationRegistry(r Registry) registrations.HandlerRegistrationRegistry[*Target, TransferOption] {
	if r == nil {
		return nil
	}
	return r.AsHandlerRegistrationRegistry()
}

func Usage(ctx ocm.Context) string {
	list := For(ctx).AsHandlerRegistrationRegistry().GetHandlers(NewTarget(ctx))
	sort.Sort(list)
	return `The following transfer hander names are supported:
` + listformat.FormatListElements("", list)
}
