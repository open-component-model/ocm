package ocireg

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/containerd/containerd/remotes"
	"github.com/containerd/errdefs"
	"github.com/containerd/log"
	"github.com/opencontainers/go-digest"
	"github.com/sirupsen/logrus"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/tech/oras"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/logging"
)

// TODO: add cache

type dataAccess struct {
	accessio.NopCloser
	lock    sync.Mutex
	fetcher remotes.Fetcher
	desc    artdesc.Descriptor
	reader  io.ReadCloser
}

var _ cpi.DataAccess = (*dataAccess)(nil)

func NewDataAccess(fetcher remotes.Fetcher, digest digest.Digest, mimeType string, delayed bool) (*dataAccess, error) {
	var reader io.ReadCloser
	var err error
	desc := artdesc.Descriptor{
		MediaType: mimeType,
		Digest:    digest,
		Size:      blobaccess.BLOB_UNKNOWN_SIZE,
	}
	if !delayed {
		reader, err = fetcher.Fetch(dummyContext, desc)
		if err != nil {
			return nil, err
		}
	}
	return &dataAccess{
		fetcher: fetcher,
		desc:    desc,
		reader:  reader,
	}, nil
}

func (d *dataAccess) Get() ([]byte, error) {
	return readAll(d.Reader())
}

func (d *dataAccess) Reader() (io.ReadCloser, error) {
	d.lock.Lock()
	reader := d.reader
	d.reader = nil
	d.lock.Unlock()
	if reader != nil {
		return reader, nil
	}
	return d.fetcher.Fetch(dummyContext, d.desc)
}

func readAll(reader io.ReadCloser, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func push(ctx context.Context, p oras.Pusher, blob blobaccess.BlobAccess) error {
	desc := *artdesc.DefaultBlobDescriptor(blob)
	return pushData(ctx, p, desc, blob)
}

func pushData(ctx context.Context, p oras.Pusher, desc artdesc.Descriptor, data blobaccess.DataAccess) error {
	key := remotes.MakeRefKey(ctx, desc)
	if desc.Size == 0 {
		desc.Size = -1
	}

	logging.Logger().Debug("*** push blob", "mediatype", desc.MediaType, "digest", desc.Digest, "key", key)
	if err := p.Push(ctx, desc, data); err != nil {
		if errdefs.IsAlreadyExists(err) {
			logging.Logger().Debug("blob already exists", "mediatype", desc.MediaType, "digest", desc.Digest)

			return nil
		}

		return fmt.Errorf("failed to push: %w", err)
	}

	return nil
}

var dummyContext = nologger()

func nologger() context.Context {
	ctx := context.Background()
	logger := logrus.New()
	logger.Level = logrus.ErrorLevel
	return log.WithLogger(ctx, logrus.NewEntry(logger))
}
