package ocireg

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/containerd/containerd/remotes"
	"github.com/containerd/log"
	"github.com/opencontainers/go-digest"
	"github.com/sirupsen/logrus"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry"

	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/logging"
)

// TODO: add cache

type dataAccess struct {
	accessio.NopCloser
	lock   sync.Mutex
	repo   registry.Repository
	desc   artdesc.Descriptor
	reader io.ReadCloser
}

var _ cpi.DataAccess = (*dataAccess)(nil)

func NewDataAccess(repo registry.Repository, digest digest.Digest, delayed bool) (*dataAccess, error) {
	var reader io.ReadCloser
	desc, err := repo.Resolve(dummyContext, digest.String())
	if err != nil {
		if errors.Is(err, errdef.ErrNotFound) {
			desc, err = repo.Blobs().Resolve(dummyContext, digest.String())
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("failed to resolve descriptor with digest %s: %w", digest.String(), err)
		}
	}
	if !delayed {
		reader, err = repo.Fetch(dummyContext, desc)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch descriptor: %w", err)
		}
	}
	return &dataAccess{
		repo:   repo,
		desc:   desc,
		reader: reader,
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
	return d.repo.Fetch(dummyContext, d.desc)
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

func push(ctx context.Context, p content.Pusher, blob blobaccess.BlobAccess) error {
	desc := *artdesc.DefaultBlobDescriptor(blob)
	return pushData(ctx, p, desc, blob)
}

func pushData(ctx context.Context, p content.Pusher, desc artdesc.Descriptor, data blobaccess.DataAccess) error {
	key := remotes.MakeRefKey(ctx, desc)
	if desc.Size == 0 {
		desc.Size = -1
	}

	logging.Logger().Debug("*** push blob", "mediatype", desc.MediaType, "digest", desc.Digest, "key", key)
	reader, err := data.Reader()
	if err != nil {
		return err
	}

	if err := p.Push(ctx, desc, reader); err != nil {
		if errors.Is(err, errdef.ErrAlreadyExists) {
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
