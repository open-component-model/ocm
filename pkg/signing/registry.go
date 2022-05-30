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

package signing

import (
	"sync"
)

type Registry interface {
	HandlerRegistry
	KeyRegistry
}

type HandlerRegistry interface {
	RegisterSignatureHandler(handler SignatureHandler)
	RegisterSigner(signer Signer)
	RegisterVerifier(verifier Verifier)
	GetSigner(name string) Signer
	GetVerifier(name string) Verifier

	RegisterHasher(hasher Hasher)
	GetHasher(name string) Hasher
}

type KeyRegistry interface {
	RegisterPublicKey(name string, key interface{})
	RegisterPrivateKey(name string, key interface{})
	GetPublicKey(name string) interface{}
	GetPrivateKey(name string) interface{}
}

////////////////////////////////////////////////////////////////////////////////

type handlerRegistry struct {
	lock     sync.RWMutex
	signers  map[string]Signer
	verifier map[string]Verifier
	hasher   map[string]Hasher
}

var _ HandlerRegistry = (*handlerRegistry)(nil)

func NewHandlerRegistry() HandlerRegistry {
	return &handlerRegistry{
		signers:  map[string]Signer{},
		verifier: map[string]Verifier{},
		hasher:   map[string]Hasher{},
	}
}

func (r *handlerRegistry) RegisterSignatureHandler(handler SignatureHandler) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.signers[handler.Algorithm()] = handler
	r.verifier[handler.Algorithm()] = handler
}

func (r *handlerRegistry) RegisterSigner(signer Signer) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.signers[signer.Algorithm()] = signer
	if v, ok := signer.(Verifier); ok && r.verifier[signer.Algorithm()] == nil {
		r.verifier[signer.Algorithm()] = v
	}
}

func (r *handlerRegistry) RegisterVerifier(verifier Verifier) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.verifier[verifier.Algorithm()] = verifier
	if v, ok := verifier.(Signer); ok && r.signers[verifier.Algorithm()] == nil {
		r.signers[verifier.Algorithm()] = v
	}
}

func (r *handlerRegistry) RegisterHasher(hasher Hasher) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.hasher[hasher.Algorithm()] = hasher
}

func (r *handlerRegistry) GetSigner(name string) Signer {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.signers[name]
}

func (r *handlerRegistry) GetVerifier(name string) Verifier {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.verifier[name]
}

func (r *handlerRegistry) GetHasher(name string) Hasher {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.hasher[name]
}

////////////////////////////////////////////////////////////////////////////////

var defaultHandlerRegistry = NewHandlerRegistry()

func DefaultHandlerRegistry() HandlerRegistry {
	return defaultHandlerRegistry
}

////////////////////////////////////////////////////////////////////////////////

type keyRegistry struct {
	lock        sync.RWMutex
	publicKeys  map[string]interface{}
	privateKeys map[string]interface{}
}

var _ KeyRegistry = (*keyRegistry)(nil)

func NewKeyRegistry() KeyRegistry {
	return &keyRegistry{
		publicKeys:  map[string]interface{}{},
		privateKeys: map[string]interface{}{},
	}
}

func (r *keyRegistry) RegisterPublicKey(name string, key interface{}) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.publicKeys[name] = key
}

func (r *keyRegistry) RegisterPrivateKey(name string, key interface{}) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.privateKeys[name] = key
}

func (r *keyRegistry) GetPublicKey(name string) interface{} {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.publicKeys[name]
}

func (r *keyRegistry) GetPrivateKey(name string) interface{} {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.privateKeys[name]
}

var defaultKeyRegistry = NewKeyRegistry()

func DefaultKeyRegistry() KeyRegistry {
	return defaultKeyRegistry
}

////////////////////////////////////////////////////////////////////////////////

type registry struct {
	baseHandlers HandlerRegistry
	baseKeys     KeyRegistry
	handlers     HandlerRegistry
	keys         KeyRegistry
}

var _ Registry = (*registry)(nil)

func NewRegistry(h HandlerRegistry, k KeyRegistry) Registry {
	return &registry{
		baseHandlers: h,
		baseKeys:     k,
		handlers:     NewHandlerRegistry(),
		keys:         NewKeyRegistry(),
	}
}

func (r *registry) RegisterSignatureHandler(handler SignatureHandler) {
	r.handlers.RegisterSignatureHandler(handler)
}

func (r *registry) RegisterSigner(signer Signer) {
	r.handlers.RegisterSigner(signer)
}

func (r *registry) RegisterVerifier(verifier Verifier) {
	r.handlers.RegisterVerifier(verifier)
}

func (r *registry) GetSigner(name string) Signer {
	s := r.handlers.GetSigner(name)
	if s == nil && r.baseHandlers != nil {
		s = r.baseHandlers.GetSigner(name)
	}
	return s
}

func (r *registry) GetVerifier(name string) Verifier {
	s := r.handlers.GetVerifier(name)
	if s == nil && r.baseHandlers != nil {
		s = r.baseHandlers.GetVerifier(name)
	}
	return s
}

func (r *registry) RegisterHasher(hasher Hasher) {
	r.handlers.RegisterHasher(hasher)
}

func (r *registry) GetHasher(name string) Hasher {
	s := r.handlers.GetHasher(name)
	if s == nil && r.baseHandlers != nil {
		s = r.baseHandlers.GetHasher(name)
	}
	return s
}

func (r *registry) RegisterPublicKey(name string, key interface{}) {
	r.keys.RegisterPublicKey(name, key)
}

func (r *registry) RegisterPrivateKey(name string, key interface{}) {
	r.keys.RegisterPrivateKey(name, key)
}

func (r *registry) GetPublicKey(name string) interface{} {
	s := r.keys.GetPublicKey(name)
	if s == nil && r.baseKeys != nil {
		s = r.baseKeys.GetPublicKey(name)
	}
	return s
}

func (r *registry) GetPrivateKey(name string) interface{} {
	s := r.keys.GetPrivateKey(name)
	if s == nil && r.baseKeys != nil {
		s = r.baseKeys.GetPrivateKey(name)
	}
	return s
}

var defaultRegistry = NewRegistry(DefaultHandlerRegistry(), DefaultKeyRegistry())

func DefaultRegistry() Registry {
	return defaultRegistry
}
