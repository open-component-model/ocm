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
	"fmt"
	"sync"

	"github.com/opencontainers/go-digest"
)

type DigesterType struct {
	Kind    string `json:"kind"`
	Version string `json:"version"`
}

func (t DigesterType) String() string {
	return fmt.Sprintf("%s/%s", t.Kind, t.Version)
}

type DigestDescriptor struct {
	Digest   digest.Digest `json:"digest"`
	Digester *DigesterType `json:"digester,omitempty"`
}

func NewDigestDescriptor(digest digest.Digest, typ DigesterType) *DigestDescriptor {
	return &DigestDescriptor{
		Digest:   digest,
		Digester: &typ,
	}
}

// BlobDigester is the interface for digest providers
// for dedicated mime types.
// If found the digest provided by the digester will
// replace the standard digest calculated for the byte content
// of the blob
type BlobDigester interface {
	GetType() DigesterType
	DetermineDigest(resType string, meth AccessMethod) (*DigestDescriptor, error)
}

// BlobDigesterRegistry registers blob handlers to use in a dedicated ocm context
type BlobDigesterRegistry interface {
	// RegisterDigester registers a blob digester for a dedicated exact mime type
	//
	RegisterDigester(handler BlobDigester, restypes ...string)
	// GetDigester returns the digester for a given type
	GetDigester(typ DigesterType) BlobDigester
	DetermineDigests(typ string, acc AccessMethod, typs ...DigesterType) ([]DigestDescriptor, error)
}

////////////////////////////////////////////////////////////////////////////////

type blobDigesterRegistry struct {
	lock         sync.RWMutex
	typehandlers map[string][]BlobDigester
	digesters    map[DigesterType]BlobDigester
}

var DefaultBlobDigesterRegistry = NewBlobDigesterRegistry()

func NewBlobDigesterRegistry() BlobDigesterRegistry {
	return &blobDigesterRegistry{typehandlers: map[string][]BlobDigester{}, digesters: map[DigesterType]BlobDigester{}}
}

func (r *blobDigesterRegistry) RegisterDigester(digester BlobDigester, restypes ...string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	old := r.digesters[digester.GetType()]
	if old != nil && old != digester {
		panic(fmt.Errorf("duplicate digester type %q: %T and %T", old.GetType(), old, digester))
	}
	r.digesters[digester.GetType()] = digester

outer:
	for _, t := range restypes {
		old := r.typehandlers[t]
		for _, o := range old {
			if o == digester {
				continue outer
			}
		}
		old = append(old, digester)
		r.typehandlers[t] = old
	}
}

func (r *blobDigesterRegistry) GetDigester(typ DigesterType) BlobDigester {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.digesters[typ]
}

func (r *blobDigesterRegistry) handle(list []BlobDigester, typ string, acc AccessMethod) ([]DigestDescriptor, error) {
	for _, h := range list {
		d, err := h.DetermineDigest(typ, acc)
		if err != nil {
			return nil, err
		}
		if d != nil {
			return []DigestDescriptor{
				*d,
			}, nil
		}
	}
	return nil, nil
}

func (r *blobDigesterRegistry) DetermineDigests(typ string, acc AccessMethod, dtyps ...DigesterType) ([]DigestDescriptor, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	if len(dtyps) == 0 {
		var err error
		res, err := r.handle(r.typehandlers[typ], typ, acc)
		if res != nil || err != nil {
			return res, err
		}
		res, err = r.handle(r.typehandlers[""], typ, acc)
		if res != nil || err != nil {
			return res, err
		}
		d, err := defaultDigester{}.DetermineDigest(typ, acc)
		if err != nil {
			return nil, err
		}
		return []DigestDescriptor{
			*d,
		}, nil
	}

	var result []DigestDescriptor
	for _, dtyp := range dtyps {
		t := r.digesters[dtyp]
		if t != nil {
			d, err := t.DetermineDigest(typ, acc)
			if err != nil {
				return nil, err
			}
			if d != nil {
				result = append(result, *d)
			}
		}
	}
	return result, nil
}

func RegisterDigester(digester BlobDigester, arttypes ...string) {
	DefaultBlobDigesterRegistry.RegisterDigester(digester, arttypes...)
}

////////////////////////////////////////////////////////////////////////////////

type defaultDigester struct{}

var _ BlobDigester = (*defaultDigester)(nil)

func (d defaultDigester) GetType() DigesterType {
	return DigesterType{
		Kind: "bytes",
	}
}

func (d defaultDigester) DetermineDigest(typ string, acc AccessMethod) (*DigestDescriptor, error) {
	r, err := acc.Reader()
	if err != nil {
		return nil, err
	}
	dig, err := digest.FromReader(r)
	if err != nil {
		return nil, err
	}
	return &DigestDescriptor{
		Digest:   dig,
		Digester: nil,
	}, nil
}
