// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package core

import (
	"strings"
	"sync"
)

// StorageContext is an object describing the storage context used for the
// mapping of a component repository to a base repository (e.g. oci api)
// It depends on the Context type of the used base repository
type StorageContext interface{}

// BlobHandler s the interface for a dedicated handling of storing blobs
// for the LocalBlob access method in a dedicated kind of repository.
// with the possibility of access by an external distribution spec.
// (besides of the blob storage as part of a component version).
// The technical repository to use should be derivable from the chosen
// component directory or passed together with the storage context.
// The task of the handler is to store the local blob on its own
// responsibility and to return an appropriate global access method.
type BlobHandler interface {
	// StoreBlob has the change to decide to store a local blob
	// in a repository specific fashion provide external access.
	// If this is possible and done an appropriate access spec
	// must be returned, if this is not done, nil has to be returned
	// without error
	StoreBlob(repo Repository, blob BlobAccess, hint string, ctx StorageContext) (AccessSpec, error)
}

// MultiBlobHandler is a BlobHandler consisting of a sequence of handlers
type MultiBlobHandler []BlobHandler

func (m MultiBlobHandler) StoreBlob(repo Repository, blob BlobAccess, hint string, ctx StorageContext) (AccessSpec, error) {
	for _, h := range m {
		a, err := h.StoreBlob(repo, blob, hint, ctx)
		if err != nil {
			return nil, err
		}
		if a != nil {
			return a, nil
		}
	}
	return nil, nil
}

// BlobHandlerKey is the registration key for BlobHandlers
type BlobHandlerKey struct {
	ContextType    string
	RepositoryType string
	MimeType       string
}

func ForRepo(ctxtype, repotype string) BlobHandlerKey {
	return BlobHandlerKey{ContextType: ctxtype, RepositoryType: repotype}
}
func ForMimeType(mimetype string) BlobHandlerKey {
	return BlobHandlerKey{ContextType: mimetype}
}

// BlobHandlerRegistry registers blob handlers to use in a dedicated ocm context
type BlobHandlerRegistry interface {
	// RegisterBlobHandler registers a blob handler. It must specify either a sole mime type,
	// or a context and repository type, or all three keys
	RegisterBlobHandler(handler BlobHandler, keys ...BlobHandlerKey)
	// GetHandler returns handler trying all matches in the following order:
	//
	// - a handle matching all keys
	//
	// - handlers matching a sole mimetype handler (from specific to more general by discarding + components)
	//
	// - a handler matching the repo
	//
	// - handlers matching everything
	GetHandler(ctxtype, repotype string, mimeType string) BlobHandler
}

type blobHandlerRegistry struct {
	lock       sync.RWMutex
	handlers   map[BlobHandlerKey]BlobHandler
	defhandler MultiBlobHandler
}

var DefaultBlobHandlerRegistry = NewBlobHandlerRegistry()

func NewBlobHandlerRegistry() BlobHandlerRegistry {
	return &blobHandlerRegistry{handlers: map[BlobHandlerKey]BlobHandler{}}
}

func (r *blobHandlerRegistry) RegisterBlobHandler(handler BlobHandler, keys ...BlobHandlerKey) {
	r.lock.Lock()
	defer r.lock.Unlock()

	key := BlobHandlerKey{}
	for _, k := range keys {
		if k.ContextType != "" {
			key.ContextType = k.ContextType
		}
		if k.RepositoryType != "" {
			key.RepositoryType = k.RepositoryType
		}
		if k.MimeType != "" {
			key.MimeType = k.MimeType
		}
	}

	def := BlobHandlerKey{}

	if key == def {
		r.defhandler = append(r.defhandler, handler)
	} else {
		r.handlers[key] = handler
	}
}

func (r *blobHandlerRegistry) forMimeType(ctxtype, repotype, mimetype string) MultiBlobHandler {
	var multi MultiBlobHandler

	mime := mimetype
	for {
		if h := r.handlers[BlobHandlerKey{ctxtype, repotype, mime}]; h != nil {
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

func (r *blobHandlerRegistry) GetHandler(ctxtype, repotype, mimetype string) BlobHandler {
	r.lock.RLock()
	defer r.lock.RUnlock()

	var multi MultiBlobHandler
	if ctxtype != "" || repotype != "" {
		multi = append(multi, r.forMimeType(ctxtype, repotype, mimetype))
	}
	multi = append(multi, r.forMimeType("", "", mimetype))
	multi = append(multi, r.defhandler)
	if len(multi) == 0 {
		return nil
	}
	return multi
}

func RegisterBlobHandler(handler BlobHandler, keys ...BlobHandlerKey) {
	DefaultBlobHandlerRegistry.RegisterBlobHandler(handler, keys...)
}
