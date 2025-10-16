package blobaccess

import (
	"bytes"
	"compress/gzip"
	"io"
	"sync"

	"github.com/mandelsoft/goutils/errors"
	"github.com/opencontainers/go-digest"
	"ocm.software/ocm/api/utils/blobaccess/bpi"
	"ocm.software/ocm/api/utils/compression"
	"ocm.software/ocm/api/utils/mime"
)

////////////////////////////////////////////////////////////////////////////////

type _compression struct {
	blob bpi.BlobAccess
}

var _ bpi.BlobAccessBase = (*_compression)(nil)

func (c *_compression) Close() error {
	return c.blob.Close()
}

func (c *_compression) Get() ([]byte, error) {
	r, err := c.blob.Reader()
	if err != nil {
		return nil, err
	}
	defer r.Close()
	rr, _, err := compression.AutoDecompress(r)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(nil)

	w := gzip.NewWriter(buf)
	_, err = io.Copy(w, rr)
	w.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type reader struct {
	wait sync.WaitGroup
	io.ReadCloser
	err error
}

func (r *reader) Close() error {
	err := r.ReadCloser.Close()
	r.wait.Wait()
	return errors.Join(err, r.err)
}

func (c *_compression) Reader() (io.ReadCloser, error) {
	r, err := c.blob.Reader()
	if err != nil {
		return nil, err
	}
	defer r.Close()
	rr, _, err := compression.AutoDecompress(r)
	if err != nil {
		return nil, err
	}
	pr, pw := io.Pipe()
	cw := gzip.NewWriter(pw)

	outr := &reader{ReadCloser: pr}
	outr.wait.Add(1)

	go func() {
		_, err := io.Copy(cw, rr)
		outr.err = errors.Join(err, cw.Close(), pw.Close())
		outr.wait.Done()
	}()
	return outr, nil
}

func (c *_compression) Digest() digest.Digest {
	return bpi.BLOB_UNKNOWN_DIGEST
}

func (c *_compression) MimeType() string {
	m := c.blob.MimeType()
	if mime.IsGZip(m) {
		return m
	}
	return m + "+gzip"
}

func (c *_compression) DigestKnown() bool {
	return false
}

func (c *_compression) Size() int64 {
	return bpi.BLOB_UNKNOWN_SIZE
}

func WithCompression(blob bpi.BlobAccess) (bpi.BlobAccess, error) {
	b, err := blob.Dup()
	if err != nil {
		return nil, err
	}
	return bpi.NewBlobAccessForBase(&_compression{
		blob: b,
	}), nil
}

////////////////////////////////////////////////////////////////////////////////

type decompression struct {
	blob bpi.BlobAccess
}

var _ bpi.BlobAccessBase = (*decompression)(nil)

func (c *decompression) Close() error {
	return c.blob.Close()
}

func (c *decompression) Get() ([]byte, error) {
	r, err := c.blob.Reader()
	if err != nil {
		return nil, err
	}
	defer r.Close()
	rr, _, err := compression.AutoDecompress(r)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, rr)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *decompression) Reader() (io.ReadCloser, error) {
	r, err := c.blob.Reader()
	if err != nil {
		return nil, err
	}
	defer r.Close()
	rr, _, err := compression.AutoDecompress(r)
	return rr, err
}

func (c *decompression) Digest() digest.Digest {
	return bpi.BLOB_UNKNOWN_DIGEST
}

func (c *decompression) MimeType() string {
	m := c.blob.MimeType()
	if !mime.IsGZip(m) {
		return m
	}
	return m[:len(m)-5]
}

func (c *decompression) DigestKnown() bool {
	return false
}

func (c *decompression) Size() int64 {
	return bpi.BLOB_UNKNOWN_SIZE
}

func WithDecompression(blob bpi.BlobAccess) (bpi.BlobAccess, error) {
	b, err := blob.Dup()
	if err != nil {
		return nil, err
	}
	return bpi.NewBlobAccessForBase(&decompression{
		blob: b,
	}), nil
}
