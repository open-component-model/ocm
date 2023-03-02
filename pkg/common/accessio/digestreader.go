// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package accessio

import (
	"crypto"
	"hash"
	"io"

	"github.com/opencontainers/go-digest"
)

// wow. digest does support a map with supported digesters. Unfortunately this one does not
// contain all the crypto hashes AND this map is private NAD there is no function to add entries,
// so that it cannot be extended from outside the package. I love GO.
// Therefore, we have to fake it a little to support digests with other crypto hashes.

type DigestReader struct {
	reader io.Reader
	alg    digest.Algorithm
	hash   hash.Hash
	count  int64
}

func (r *DigestReader) Size() int64 {
	return r.count
}

func (r *DigestReader) Digest() digest.Digest {
	return digest.NewDigest(r.alg, r.hash)
}

func (r *DigestReader) Read(buf []byte) (int, error) {
	c, err := r.reader.Read(buf)
	if c > 0 {
		r.count += int64(c)
		r.hash.Write(buf[:c])
	}
	return c, err
}

func NewDefaultDigestReader(r io.Reader) *DigestReader {
	return NewDigestReaderWith(digest.Canonical, r)
}

func NewDigestReaderWith(algorithm digest.Algorithm, r io.Reader) *DigestReader {
	digester := algorithm.Digester()
	return &DigestReader{
		reader: r,
		hash:   digester.Hash(),
		alg:    algorithm,
		count:  0,
	}
}

func NewDigestReaderWithHash(hash crypto.Hash, r io.Reader) *DigestReader {
	return &DigestReader{
		reader: r,
		hash:   hash.New(),
		alg:    digest.Algorithm(hash.String()), // fake a non-supported digest algorithm
		count:  0,
	}
}

func Digest(access DataAccess) (digest.Digest, error) {
	reader, err := access.Reader()
	if err != nil {
		return "", err
	}
	defer reader.Close()

	dig, err := digest.FromReader(reader)
	if err != nil {
		return "", err
	}
	return dig, nil
}
