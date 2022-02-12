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

	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/errors"
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
	DetermineDigest(blob accessio.BlobAccess) (*DigestDescriptor, error)
}

// BlobDigesterRegistry registers blob handlers to use in a dedicated ocm context
type BlobDigesterRegistry interface {
	// RegisterDigester registers a blob digester for a dedicated exact mime type
	//
	RegisterDigester(handler BlobDigester, mimetypes ...string)
	// GetDigester returns the digester for a given type
	GetDigester(typ DigesterType) BlobDigester
	DetermineDigests(blob accessio.BlobAccess, typs ...DigesterType) ([]DigestDescriptor, error)
}

type blobDigesterRegistry struct {
	lock         sync.RWMutex
	mimehandlers map[string]BlobDigester
	digesters    map[DigesterType]BlobDigester
}

var DefaultBlobDigesterRegistry = NewBlobDigesterRegistry()

func NewBlobDigesterRegistry() BlobDigesterRegistry {
	return &blobDigesterRegistry{mimehandlers: map[string]BlobDigester{}, digesters: map[DigesterType]BlobDigester{}}
}

func (r *blobDigesterRegistry) RegisterDigester(digester BlobDigester, mimetypes ...string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	old := r.digesters[digester.GetType()]
	if old != nil && old != digester {
		panic(fmt.Errorf("duplicate digester type %q: %T and %T", old.GetType(), old, digester))
	}
	r.digesters[digester.GetType()] = digester

	for _, mimetype := range mimetypes {
		old = r.mimehandlers[mimetype]
		if old != nil && old != digester {
			panic(fmt.Errorf("duplicate digester for mime type %q: %s and %s", mimetype, old.GetType(), digester.GetType()))
		}
		r.mimehandlers[mimetype] = digester
	}
}

func (r *blobDigesterRegistry) GetDigester(typ DigesterType) BlobDigester {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.digesters[typ]
}

func (r *blobDigesterRegistry) DetermineDigests(blob accessio.BlobAccess, typs ...DigesterType) ([]DigestDescriptor, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	if len(typs) == 0 {
		var d *DigestDescriptor
		var err error
		h := r.mimehandlers[blob.MimeType()]
		if h != nil {
			d, err = h.DetermineDigest(blob)
		} else {
			d, err = defaultDigester{}.DetermineDigest(blob)
		}
		if err != nil {
			return nil, err
		}
		return []DigestDescriptor{
			*d,
		}, nil
	}
	var result []DigestDescriptor
	for _, typ := range typs {
		t := r.digesters[typ]
		if t != nil {
			d, err := t.DetermineDigest(blob)
			if err != nil {
				return nil, err
			}
			result = append(result, *d)
		}
	}
	return result, nil
}

func RegisterDigester(digester BlobDigester, mimetypes ...string) {
	DefaultBlobDigesterRegistry.RegisterDigester(digester, mimetypes...)
}

////////////////////////////////////////////////////////////////////////////////

type defaultDigester struct{}

var _ BlobDigester = (*defaultDigester)(nil)

func (d defaultDigester) GetType() DigesterType {
	return DigesterType{
		Kind: "bytes",
	}
}

func (d defaultDigester) DetermineDigest(blob accessio.BlobAccess) (*DigestDescriptor, error) {
	digest := blob.Digest()
	if digest == "" {
		return nil, errors.New("no digest available")
	}
	return &DigestDescriptor{
		Digest:   digest,
		Digester: nil,
	}, nil
}
