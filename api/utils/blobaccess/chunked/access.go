package chunked

import (
	"io"
	"sync"

	"github.com/opencontainers/go-digest"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/blobaccess/bpi"
)

type Chunked interface {
	bpi.BlobAccess

	Next() bool
}

type chunked struct {
	lock sync.Mutex
	base bpi.BlobAccess
	blobsize uint64
	chunksize uint64
	preread uint

	reader io.Reader
}

var _ bpi.BlobAccessBase = (*chunked)(nil)

func New(acc bpi.BlobAccess, chunk uint64, preread...uint) (Chunked, error) {
	b, err := acc.Dup()
	if err != nil {
		return nil, err
	}

	s := acc.Size()

	return bpi.NewBlobAccessForBase(&chunked{base: b, blobsize: size, chunksize: chunk, preread: utils.OptionalDefaulted(8096, preread...)}), nil
}

type view struct {
	bpi.BlobAccess
}

func (v *view) Dup() bpi.BlobAccess {

}
func (c *chunked) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.base == nil {
		return bpi.ErrClosed
	}
	err := c.base.Close()
	c.base = nil
	return err
}

func (c *chunked) Get() ([]byte, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if
	// TODO implement me
	panic("implement me")
}

func (c *chunked) Reader() (io.ReadCloser, error) {
	// TODO implement me
	panic("implement me")
}

func (c *chunked) Digest() digest.Digest {
	// TODO implement me
	panic("implement me")
}

func (c *chunked) MimeType() string {
	// TODO implement me
	panic("implement me")
}

func (c *chunked) DigestKnown() bool {
	// TODO implement me
	panic("implement me")
}

func (c *chunked) Size() int64 {
	// TODO implement me
	panic("implement me")
}
