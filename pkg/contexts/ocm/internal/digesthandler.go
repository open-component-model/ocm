// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"fmt"
	"io"
	"sync"

	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/signing"
)

type DigesterType struct {
	HashAlgorithm          string
	NormalizationAlgorithm string
}

type DigestDescriptor = metav1.DigestSpec

func NewDigestDescriptor(digest, hashAlgo, normAlgo string) *DigestDescriptor {
	return &DigestDescriptor{
		HashAlgorithm:          hashAlgo,
		NormalisationAlgorithm: normAlgo,
		Value:                  digest,
	}
}

// BlobDigester is the interface for digest providers
// for dedicated mime types.
// If found the digest provided by the digester will
// replace the standard digest calculated for the byte content
// of the blob.
type BlobDigester interface {
	GetType() DigesterType
	DetermineDigest(resType string, meth AccessMethod, preferred signing.Hasher) (*DigestDescriptor, error)
}

// BlobDigesterRegistry registers blob handlers to use in a dedicated ocm context.
type BlobDigesterRegistry interface {
	IsInitial() bool
	// MustRegisterDigester registers a blob digester for a dedicated exact mime type
	//
	Register(handler BlobDigester, restypes ...string) error
	// GetDigester returns the digester for a given type
	GetDigester(typ DigesterType) BlobDigester

	GetDigesterForType(t string) []BlobDigester
	DetermineDigests(typ string, preferred signing.Hasher, registry signing.Registry, acc AccessMethod, typs ...DigesterType) ([]DigestDescriptor, error)

	Copy() BlobDigesterRegistry
}

////////////////////////////////////////////////////////////////////////////////

type blobDigesterRegistry struct {
	lock         sync.RWMutex
	typehandlers map[string][]BlobDigester
	normhandlers map[string][]BlobDigester
	digesters    map[DigesterType]BlobDigester
}

var DefaultBlobDigesterRegistry = NewBlobDigesterRegistry()

func NewBlobDigesterRegistry() BlobDigesterRegistry {
	return &blobDigesterRegistry{
		typehandlers: map[string][]BlobDigester{},
		normhandlers: map[string][]BlobDigester{},
		digesters:    map[DigesterType]BlobDigester{},
	}
}

func (r *blobDigesterRegistry) IsInitial() bool {
	return len(r.typehandlers) == 0 && len(r.normhandlers) == 0 && len(r.digesters) == 0
}

func (r *blobDigesterRegistry) Register(digester BlobDigester, restypes ...string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	t := digester.GetType()
	old := r.digesters[t]
	if old != nil && old != digester {
		return fmt.Errorf("duplicate digester type %q: %T and %T", t, old, digester)
	}
	r.digesters[t] = digester

	oldn := r.normhandlers[t.NormalizationAlgorithm]
outer_norm:
	for _, o := range oldn {
		if o == digester {
			continue outer_norm
		}
	}
	oldn = append(oldn, digester)
	r.normhandlers[t.NormalizationAlgorithm] = oldn

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
	return nil
}

func (r *blobDigesterRegistry) GetDigester(typ DigesterType) BlobDigester {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.digesters[typ]
}

func (r *blobDigesterRegistry) GetDigesterForType(typ string) []BlobDigester {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return append(r.typehandlers[typ][:0:0], r.typehandlers[typ]...)
}

func (r *blobDigesterRegistry) Copy() BlobDigesterRegistry {
	r.lock.RLock()
	defer r.lock.RUnlock()

	n := NewBlobDigesterRegistry().(*blobDigesterRegistry)
	for k, v := range r.typehandlers {
		n.typehandlers[k] = append(v[:0:0], v...)
	}
	for k, v := range r.normhandlers {
		n.normhandlers[k] = append(v[:0:0], v...)
	}
	for k, v := range r.digesters {
		n.digesters[k] = v
	}
	return n
}

func (r *blobDigesterRegistry) handle(list []BlobDigester, typ string, acc AccessMethod, preferred signing.Hasher) ([]DigestDescriptor, error) {
	for _, h := range list {
		d, err := h.DetermineDigest(typ, acc, preferred)
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

func (r *blobDigesterRegistry) DetermineDigests(restype string, preferred signing.Hasher, registry signing.Registry, acc AccessMethod, dtyps ...DigesterType) ([]DigestDescriptor, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	none := DigesterType{}
	for i := 0; i < len(dtyps); i++ {
		if dtyps[i] == none {
			dtyps = append(dtyps[:i], dtyps[i+1:]...)
			i--
		}
	}
	if len(dtyps) == 0 {
		var err error
		res, err := r.handle(r.typehandlers[restype], restype, acc, preferred)
		if res != nil || err != nil {
			return res, err
		}
		res, err = r.handle(r.typehandlers[""], restype, acc, preferred)
		if res != nil || err != nil {
			return res, err
		}
		d, err := defaultDigester{}.DetermineDigest(restype, acc, preferred)
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
			d, err := t.DetermineDigest(restype, acc, preferred)
			if err != nil {
				return nil, err
			}
			if d != nil {
				result = append(result, *d)
			}
		}
	}
	if len(result) == 0 {
		for _, dtyp := range dtyps {
			if dtyp.NormalizationAlgorithm != "" {
				hasher := preferred
				if dtyp.HashAlgorithm != "" {
					hasher = registry.GetHasher(dtyp.HashAlgorithm)
				}
				if hasher == nil {
					continue
				}
				for _, t := range r.normhandlers[dtyp.NormalizationAlgorithm] {
					d, err := t.DetermineDigest(restype, acc, hasher)
					if err != nil {
						return nil, err
					}
					if d != nil {
						result = append(result, *d)
						continue
					}
				}
			}
		}
	}
	return result, nil
}

func MustRegisterDigester(digester BlobDigester, arttypes ...string) {
	err := DefaultBlobDigesterRegistry.Register(digester, arttypes...)
	if err != nil {
		panic(err)
	}
}

////////////////////////////////////////////////////////////////////////////////

const GenericBlobDigestV1 = "genericBlobDigest/v1"

func init() {
	MustRegisterDigester(&defaultDigester{})
}

type defaultDigester struct{}

var _ BlobDigester = (*defaultDigester)(nil)

func (d defaultDigester) GetType() DigesterType {
	return DigesterType{
		HashAlgorithm:          "",
		NormalizationAlgorithm: GenericBlobDigestV1,
	}
}

func (d defaultDigester) DetermineDigest(typ string, acc AccessMethod, preferred signing.Hasher) (*DigestDescriptor, error) {
	r, err := acc.Reader()
	if err != nil {
		return nil, err
	}
	hash := preferred.Create()

	if _, err := io.Copy(hash, r); err != nil {
		return nil, err
	}

	return &DigestDescriptor{
		Value:                  fmt.Sprintf("%x", hash.Sum(nil)),
		HashAlgorithm:          preferred.Algorithm(),
		NormalisationAlgorithm: GenericBlobDigestV1,
	}, nil
}
