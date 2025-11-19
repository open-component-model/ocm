package ocireg

import (
	"sync"

	"github.com/containerd/containerd/remotes"
	"github.com/mandelsoft/goutils/errors"
	"github.com/opencontainers/go-digest"
	"github.com/sirupsen/logrus"

	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/oci/extensions/attrs/cacheattr"
	"ocm.software/ocm/api/tech/oras"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
)

type BlobContainer interface {
	GetBlobData(digest digest.Digest) (int64, cpi.DataAccess, error)
	AddBlob(blob cpi.BlobAccess) (int64, digest.Digest, error)
	Unref() error
}

type blobContainer struct {
	accessio.StaticAllocatable
	fetcher oras.Fetcher
	pusher  oras.Pusher
	mime    string
}

type BlobContainers struct {
	cache   accessio.BlobCache
	fetcher oras.Fetcher
	pusher  oras.Pusher

	mimes sync.Map // map[string]BlobContainer
}

func NewBlobContainers(ctx cpi.Context, fetcher remotes.Fetcher, pusher oras.Pusher) *BlobContainers {
	return &BlobContainers{
		cache:   cacheattr.Get(ctx),
		fetcher: fetcher,
		pusher:  pusher,
	}
}

func (c *BlobContainers) Get(mime string) (BlobContainer, error) {
	// Fast path: load existing
	if v, ok := c.mimes.Load(mime); ok {
		return v.(BlobContainer), nil
	}

	// Slow path: need to create a new one
	newBC, err := NewBlobContainer(c.cache, mime, c.fetcher, c.pusher)
	if err != nil {
		return nil, err
	}

	// Try to publish it. Another goroutine may win the race.
	actual, loaded := c.mimes.LoadOrStore(mime, newBC)
	if loaded {
		// We lost the race — discard our new instance
		return actual.(BlobContainer), newBC.Unref()
	}

	return newBC, nil
}

func (c *BlobContainers) Release() error {
	list := errors.ErrListf("releasing mime block caches")

	c.mimes.Range(func(_, value any) bool {
		list.Add(value.(BlobContainer).Unref())
		return true
	})

	return list.Result()
}

func newBlobContainer(mime string, fetcher oras.Fetcher, pusher oras.Pusher) *blobContainer {
	return &blobContainer{
		mime:    mime,
		fetcher: fetcher,
		pusher:  pusher,
	}
}

func NewBlobContainer(cache accessio.BlobCache, mime string, fetcher oras.Fetcher, pusher oras.Pusher) (BlobContainer, error) {
	c := newBlobContainer(mime, fetcher, pusher)

	if cache == nil {
		return c, nil
	}
	r, err := accessio.CachedAccess(c, c, cache)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (n *blobContainer) GetBlobData(digest digest.Digest) (int64, cpi.DataAccess, error) {
	logrus.Debugf("orig get %s %s\n", n.mime, digest)
	acc, err := NewDataAccess(n.fetcher, digest, n.mime, false)
	return blobaccess.BLOB_UNKNOWN_SIZE, acc, err
}

func (n *blobContainer) AddBlob(blob cpi.BlobAccess) (int64, digest.Digest, error) {
	err := push(dummyContext, n.pusher, blob)
	if err != nil {
		return blobaccess.BLOB_UNKNOWN_SIZE, blobaccess.BLOB_UNKNOWN_DIGEST, err
	}
	return blob.Size(), blob.Digest(), err
}
