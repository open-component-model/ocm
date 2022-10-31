// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

type ImplementationRepositoryType struct {
	ContextType    string
	RepositoryType string
}

func (t ImplementationRepositoryType) String() string {
	return fmt.Sprintf("%s[%s]", t.RepositoryType, t.ContextType)
}

func (t ImplementationRepositoryType) IsInitial() bool {
	return t.RepositoryType == "" && t.ContextType == ""
}

// StorageContext is an object describing the storage context used for the
// mapping of a component repository to a base repository (e.g. oci api)
// It depends on the Context type of the used base repository.
type StorageContext interface {
	GetContext() Context
	TargetComponentVersion() ComponentVersionAccess
	TargetComponentRepository() Repository
	GetImplementationRepositoryType() ImplementationRepositoryType
}

// BlobHandler s the interface for a dedicated handling of storing blobs
// for the LocalBlob access method in a dedicated kind of repository.
// with the possibility of access by an external distribution spec.
// (besides of the blob storage as part of a component version).
// The technical repository to use should be derivable from the chosen
// component directory or passed together with the storage context.
// The task of the handler is to store the local blob on its own
// responsibility and to return an appropriate global access method.
type BlobHandler interface {
	// StoreBlob has the chance to decide to store a local blob
	// in a repository specific fashion providing external access.
	// If this is possible and done an appropriate access spec
	// must be returned, if this is not done, nil has to be returned
	// without error
	StoreBlob(blob BlobAccess, hint string, global AccessSpec, ctx StorageContext) (AccessSpec, error)
}

// MultiBlobHandler is a BlobHandler consisting of a sequence of handlers.
type MultiBlobHandler []BlobHandler

var _ sort.Interface = MultiBlobHandler(nil)

func (m MultiBlobHandler) StoreBlob(blob BlobAccess, hint string, global AccessSpec, ctx StorageContext) (AccessSpec, error) {
	for _, h := range m {
		a, err := h.StoreBlob(blob, hint, global, ctx)
		if err != nil {
			return nil, err
		}
		if a != nil {
			return a, nil
		}
	}
	return nil, nil
}

func (m MultiBlobHandler) Len() int {
	return len(m)
}

func (m MultiBlobHandler) Less(i, j int) bool {
	pi := DEFAULT_BLOBHANDLER_PRIO
	pj := DEFAULT_BLOBHANDLER_PRIO

	if p, ok := m[i].(*PrioBlobHandler); ok {
		pi = p.Prio
	}
	if p, ok := m[j].(*PrioBlobHandler); ok {
		pj = p.Prio
	}
	return pi > pj
}

func (m MultiBlobHandler) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

////////////////////////////////////////////////////////////////////////////////

type BlobHandlerOptions struct {
	BlobHandlerKey
	Priority int
}

type BlobHandlerOption interface {
	ApplyBlobHandlerOptionTo(*BlobHandlerOptions)
}

type prio struct {
	prio int
}

func WithPrio(p int) BlobHandlerOption {
	return prio{p}
}

func (o prio) ApplyBlobHandlerOptionTo(opts *BlobHandlerOptions) {
	opts.Priority = o.prio
}

////////////////////////////////////////////////////////////////////////////////

// BlobHandlerKey is the registration key for BlobHandlers.
type BlobHandlerKey struct {
	ImplementationRepositoryType
	ArtefactType string
	MimeType     string
}

var _ BlobHandlerOption = BlobHandlerKey{}

func NewBlobHandlerKey(ctxtype, repotype, artefactType, mimetype string) BlobHandlerKey {
	return BlobHandlerKey{
		ImplementationRepositoryType: ImplementationRepositoryType{
			ContextType:    ctxtype,
			RepositoryType: repotype,
		},
		ArtefactType: artefactType,
		MimeType:     mimetype,
	}
}

func (k BlobHandlerKey) ApplyBlobHandlerOptionTo(opts *BlobHandlerOptions) {
	if k.ContextType != "" {
		opts.ContextType = k.ContextType
	}
	if k.RepositoryType != "" {
		opts.RepositoryType = k.RepositoryType
	}
	if k.ArtefactType != "" {
		opts.ArtefactType = k.ArtefactType
	}
	if k.MimeType != "" {
		opts.MimeType = k.MimeType
	}
}

func ForRepo(ctxtype, repotype string) BlobHandlerOption {
	return BlobHandlerKey{ImplementationRepositoryType: ImplementationRepositoryType{ContextType: ctxtype, RepositoryType: repotype}}
}

func ForMimeType(mimetype string) BlobHandlerOption {
	return BlobHandlerKey{MimeType: mimetype}
}

func ForArtefactType(artefacttype string) BlobHandlerOption {
	return BlobHandlerKey{ArtefactType: artefacttype}
}

////////////////////////////////////////////////////////////////////////////////

// BlobHandlerRegistry registers blob handlers to use in a dedicated ocm context.
type BlobHandlerRegistry interface {
	IsInitial() bool

	// Copy provides a new independend copy of the registry
	Copy() BlobHandlerRegistry
	// RegisterBlobHandler registers a blob handler. It must specify either a sole mime type,
	// or a context and repository type, or all three keys
	Register(handler BlobHandler, opts ...BlobHandlerOption) BlobHandlerRegistry
	// GetHandler returns handler trying all matches in the following order:
	//
	// - a handler matching all keys
	// - handlers matching the repo and mime type (from specific to more general by discarding + components)
	//   - with artefact type
	//   - without artefact type
	// - handlers matching artefact type
	// - handlers matching a sole mimetype handler (from specific to more general by discarding + components)
	// - a handler matching the repo
	//
	GetHandler(repotype ImplementationRepositoryType, artefacttype, mimeType string) BlobHandler
}

const DEFAULT_BLOBHANDLER_PRIO = 100

type PrioBlobHandler struct {
	BlobHandler
	Prio int
}

type handlerCache struct {
	cache map[BlobHandlerKey]BlobHandler
}

func newHandlerCache() *handlerCache {
	return &handlerCache{map[BlobHandlerKey]BlobHandler{}}
}

func (c *handlerCache) len() int {
	return len(c.cache)
}

func (c *handlerCache) get(key BlobHandlerKey) (BlobHandler, bool) {
	h, ok := c.cache[key]
	return h, ok
}

func (c *handlerCache) set(key BlobHandlerKey, h BlobHandler) {
	c.cache[key] = h
}

type blobHandlerRegistry struct {
	lock       sync.RWMutex
	handlers   map[BlobHandlerKey]BlobHandler
	defhandler MultiBlobHandler
	cache      *handlerCache
}

var DefaultBlobHandlerRegistry = NewBlobHandlerRegistry()

func NewBlobHandlerRegistry() BlobHandlerRegistry {
	return &blobHandlerRegistry{handlers: map[BlobHandlerKey]BlobHandler{}, cache: newHandlerCache()}
}

func (r *blobHandlerRegistry) Copy() BlobHandlerRegistry {
	r.lock.RLock()
	defer r.lock.RUnlock()
	n := NewBlobHandlerRegistry().(*blobHandlerRegistry)
	n.defhandler = append(n.defhandler, r.defhandler...)
	for k, h := range r.handlers {
		n.handlers[k] = h
	}
	return n
}

func (r *blobHandlerRegistry) IsInitial() bool {
	return len(r.handlers) == 0 && len(r.defhandler) == 0
}

func (r *blobHandlerRegistry) Register(handler BlobHandler, olist ...BlobHandlerOption) BlobHandlerRegistry {
	opts := &BlobHandlerOptions{}
	for _, o := range olist {
		o.ApplyBlobHandlerOptionTo(opts)
	}
	r.lock.Lock()
	defer r.lock.Unlock()

	def := BlobHandlerKey{}

	if opts.Priority != 0 {
		handler = &PrioBlobHandler{handler, opts.Priority}
	}
	if opts.BlobHandlerKey == def {
		r.defhandler = append(r.defhandler, handler)
	} else {
		r.handlers[opts.BlobHandlerKey] = handler
	}
	if r.cache.len() > 0 {
		r.cache = newHandlerCache()
	}
	return r
}

func (r *blobHandlerRegistry) forMimeType(ctxtype, repotype, artefacttype, mimetype string) MultiBlobHandler {
	var multi MultiBlobHandler

	mime := mimetype
	for {
		if h := r.handlers[NewBlobHandlerKey(ctxtype, repotype, artefacttype, mime)]; h != nil {
			multi = append(multi, h)
		}
		idx := strings.LastIndex(mime, "+")
		if idx < 0 {
			break
		}
		mime = mime[:idx]
	}
	return multi
}

func (r *blobHandlerRegistry) GetHandler(repotype ImplementationRepositoryType, artefacttype, mimetype string) BlobHandler {
	key := BlobHandlerKey{
		ImplementationRepositoryType: repotype,
		ArtefactType:                 artefacttype,
		MimeType:                     mimetype,
	}
	h, cache := r.getHandler(key)
	if cache != nil {
		r.lock.Lock()
		defer r.lock.Unlock()
		// fill cache, if unchanged during pseudo lock upgrade (no support in go sync package for that).
		// if cache has been renewed in the meantime, just use the old outdated result, but don't update.
		if r.cache == cache {
			r.cache.set(key, h)
		}
	}
	return h
}

func (r *blobHandlerRegistry) getHandler(key BlobHandlerKey) (BlobHandler, *handlerCache) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	if h, ok := r.cache.get(key); ok {
		return h, nil
	}
	var multi MultiBlobHandler
	if !key.ImplementationRepositoryType.IsInitial() {
		multi = append(multi, r.forMimeType(key.ContextType, key.RepositoryType, key.ArtefactType, key.MimeType)...)
		if key.MimeType != "" {
			multi = append(multi, r.forMimeType(key.ContextType, key.RepositoryType, key.ArtefactType, "")...)
		}
		if key.ArtefactType != "" {
			multi = append(multi, r.forMimeType(key.ContextType, key.RepositoryType, "", key.MimeType)...)
		}
	}
	multi = append(multi, r.forMimeType("", "", key.ArtefactType, key.MimeType)...)
	if key.MimeType != "" {
		multi = append(multi, r.forMimeType("", "", key.ArtefactType, "")...)
	}
	if key.ArtefactType != "" {
		multi = append(multi, r.forMimeType("", "", "", key.MimeType)...)
	}
	if !key.ImplementationRepositoryType.IsInitial() && key.ArtefactType != "" && key.MimeType != "" {
		multi = append(multi, r.forMimeType(key.ContextType, key.RepositoryType, "", "")...)
	}
	multi = append(multi, r.defhandler...)
	if len(multi) == 0 {
		return nil, r.cache
	}
	sort.Sort(multi)
	return multi, r.cache
}

func RegisterBlobHandler(handler BlobHandler, opts ...BlobHandlerOption) {
	DefaultBlobHandlerRegistry.Register(handler, opts...)
}

func MustRegisterBlobHandler(handler BlobHandler, opts ...BlobHandlerOption) {
	DefaultBlobHandlerRegistry.Register(handler, opts...)
}
