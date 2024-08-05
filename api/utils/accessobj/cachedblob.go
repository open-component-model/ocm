package accessobj

import (
	"fmt"
	"io"
	"sync"

	"github.com/opencontainers/go-digest"

	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/datacontext/attrs/tmpcache"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/blobaccess/bpi"
	"ocm.software/ocm/api/utils/blobaccess/file"
)

type CachedBlobAccess struct {
	lock sync.Mutex
	mime string

	cache     *tmpcache.Attribute
	path      string
	digest    digest.Digest
	size      int64
	source    accessio.DataWriter
	effective blobaccess.BlobAccess
}

var _ bpi.BlobAccessBase = (*CachedBlobAccess)(nil)

func CachedBlobAccessForWriter(ctx datacontext.Context, mime string, src accessio.DataWriter) blobaccess.BlobAccess {
	return CachedBlobAccessForWriterWithCache(tmpcache.Get(ctx), mime, src)
}

func CachedBlobAccessForWriterWithCache(cache *tmpcache.Attribute, mime string, src accessio.DataWriter) blobaccess.BlobAccess {
	return bpi.NewBlobAccessForBase(&CachedBlobAccess{
		source: src,
		mime:   mime,
		cache:  cache,
	})
}

func CachedBlobAccessForDataAccess(ctx datacontext.Context, mime string, src blobaccess.DataAccess) blobaccess.BlobAccess {
	return CachedBlobAccessForWriter(ctx, mime, accessio.NewDataAccessWriter(src))
}

func (c *CachedBlobAccess) setup() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.effective != nil {
		return nil
	}

	f, err := c.cache.CreateTempFile("blob*")
	if err != nil {
		return fmt.Errorf("unable to create temporary file: %w", err)
	}
	defer f.Close()

	c.path = f.Name()

	c.size, c.digest, err = c.source.WriteTo(f)
	if err != nil {
		defer c.cache.Filesystem.Remove(f.Name())
		return fmt.Errorf("unable to write source to file '%s': %w", f.Name(), err)
	}

	c.effective = file.BlobAccess(c.mime, c.path, c.cache.Filesystem)

	return nil
}

func (c *CachedBlobAccess) Get() ([]byte, error) {
	err := c.setup()
	if err != nil {
		return nil, err
	}
	return c.effective.Get()
}

func (c *CachedBlobAccess) Reader() (io.ReadCloser, error) {
	err := c.setup()
	if err != nil {
		return nil, err
	}
	return c.effective.Reader()
}

func (c *CachedBlobAccess) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	var err error

	if c.effective != nil {
		c.effective.Close()

		err = c.cache.Filesystem.Remove(c.path)
	}

	c.effective = nil

	if err != nil {
		return fmt.Errorf("failed to close blob access cache: %w", err)
	}

	return nil
}

func (c *CachedBlobAccess) Digest() digest.Digest {
	err := c.setup()
	if err != nil {
		return bpi.BLOB_UNKNOWN_DIGEST
	}
	if c.digest == bpi.BLOB_UNKNOWN_DIGEST {
		return c.effective.Digest()
	}
	return c.digest
}

func (c *CachedBlobAccess) MimeType() string {
	return c.mime
}

func (c *CachedBlobAccess) DigestKnown() bool {
	return c.effective != nil
}

func (c *CachedBlobAccess) Size() int64 {
	err := c.setup()
	if err != nil {
		return blobaccess.BLOB_UNKNOWN_SIZE
	}
	if c.size == blobaccess.BLOB_UNKNOWN_SIZE {
		return c.effective.Size()
	}
	return c.size
}
