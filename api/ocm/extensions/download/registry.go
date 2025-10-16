package download

import (
	"sort"
	"sync"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/ocmutils/registry"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/registrations"
	"ocm.software/ocm/api/utils/runtimefinalizer"
)

const ALL = "*"

type Handler interface {
	Download(p common.Printer, racc cpi.ResourceAccess, path string, fs vfs.FileSystem) (bool, string, error)
}

const DEFAULT_BLOBHANDLER_PRIO = 100

type PrioHandler struct {
	Handler
	Prio int
}

// MultiHandler is a Handler consisting of a sequence of handlers.
type MultiHandler []Handler

var _ sort.Interface = MultiHandler(nil)

func (m MultiHandler) Download(p common.Printer, racc cpi.ResourceAccess, path string, fs vfs.FileSystem) (bool, string, error) {
	errs := errors.ErrListf("download")
	for _, h := range m {
		ok, p, err := h.Download(p, racc, path, fs)
		if ok {
			return ok, p, err
		}
		errs.Add(err)
	}
	return false, "", errs.Result()
}

func (m MultiHandler) Len() int {
	return len(m)
}

func (m MultiHandler) Less(i, j int) bool {
	pi := DEFAULT_BLOBHANDLER_PRIO
	pj := DEFAULT_BLOBHANDLER_PRIO

	if p, ok := m[i].(*PrioHandler); ok {
		pi = p.Prio
	}
	if p, ok := m[j].(*PrioHandler); ok {
		pj = p.Prio
	}
	return pi > pj
}

func (m MultiHandler) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

type Registry interface {
	Copy() Registry
	AsHandlerRegistrationRegistry() registrations.HandlerRegistrationRegistry[Target, HandlerOption]

	registrations.HandlerRegistrationRegistryAccess[Target, HandlerOption]

	Register(hdlr Handler, olist ...HandlerOption)
	LookupHandler(art, media string) MultiHandler
	Handler
	DownloadAsBlob(p common.Printer, racc cpi.ResourceAccess, path string, fs vfs.FileSystem) (bool, string, error)
}

func AsHandlerRegistrationRegistry(r Registry) registrations.HandlerRegistrationRegistry[Target, HandlerOption] {
	if r == nil {
		return nil
	}
	return r.AsHandlerRegistrationRegistry()
}

type _registry struct {
	registrations.HandlerRegistrationRegistry[Target, HandlerOption]

	id       runtimefinalizer.ObjectIdentity
	lock     sync.RWMutex
	base     Registry
	handlers *registry.Registry[Handler, registry.RegistrationKey]
}

func NewRegistry(base ...Registry) Registry {
	b := general.Optional(base...)
	return &_registry{
		id:                          runtimefinalizer.NewObjectIdentity("downloader.registry.ocm.software"),
		base:                        b,
		HandlerRegistrationRegistry: NewHandlerRegistrationRegistry(AsHandlerRegistrationRegistry(b)),
		handlers:                    registry.NewRegistry[Handler, registry.RegistrationKey](),
	}
}

func (r *_registry) AsHandlerRegistrationRegistry() registrations.HandlerRegistrationRegistry[Target, HandlerOption] {
	return r.HandlerRegistrationRegistry
}

func (r *_registry) Copy() Registry {
	n := NewRegistry(r.base).(*_registry)
	n.handlers = r.handlers.Copy()
	return n
}

func (r *_registry) LookupHandler(art, media string) MultiHandler {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.getHandlers(art, media)
}

func (r *_registry) Register(hdlr Handler, olist ...HandlerOption) {
	opts := NewHandlerOptions(olist...)
	r.lock.Lock()
	defer r.lock.Unlock()
	if opts.Priority != 0 {
		hdlr = &PrioHandler{hdlr, opts.Priority}
	}
	r.handlers.Register(registry.RegistrationKey{opts.ArtifactType, opts.MimeType}, hdlr)
}

func (r *_registry) getHandlers(arttype, mediatype string) MultiHandler {
	list := r.handlers.LookupHandler(registry.RegistrationKey{arttype, mediatype})
	if r.base != nil {
		list = append(list, r.base.LookupHandler(arttype, mediatype)...)
	}
	return list
}

func (r *_registry) Download(p common.Printer, racc cpi.ResourceAccess, path string, fs vfs.FileSystem) (bool, string, error) {
	p = common.AssurePrinter(p)
	art := racc.Meta().GetType()
	m, err := racc.AccessMethod()
	if err != nil {
		return false, "", err
	}
	defer m.Close()
	mime := m.MimeType()
	if ok, p, err := r.download(r.LookupHandler(art, mime), p, racc, path, fs); ok {
		return ok, p, err
	}
	return r.download(r.LookupHandler(ALL, ""), p, racc, path, fs)
}

func (r *_registry) DownloadAsBlob(p common.Printer, racc cpi.ResourceAccess, path string, fs vfs.FileSystem) (bool, string, error) {
	return r.download(r.LookupHandler(ALL, ""), p, racc, path, fs)
}

func (r *_registry) download(list MultiHandler, p common.Printer, racc cpi.ResourceAccess, path string, fs vfs.FileSystem) (bool, string, error) {
	sort.Stable(list)
	return list.Download(p, racc, path, fs)
}

var DefaultRegistry = NewRegistry()

func Register(hdlr Handler, olist ...HandlerOption) {
	DefaultRegistry.Register(hdlr, olist...)
}

////////////////////////////////////////////////////////////////////////////////

const ATTR_DOWNLOADER_HANDLERS = "ocm.software/ocm/api/ocm/extensions/download"

func For(ctx cpi.ContextProvider) Registry {
	if ctx == nil {
		return DefaultRegistry
	}
	return ctx.OCMContext().GetAttributes().GetOrCreateAttribute(ATTR_DOWNLOADER_HANDLERS, create).(Registry)
}

func create(datacontext.Context) interface{} {
	return NewRegistry(DefaultRegistry)
}

func SetFor(ctx datacontext.Context, registry Registry) {
	ctx.GetAttributes().SetAttribute(ATTR_DOWNLOADER_HANDLERS, registry)
}
