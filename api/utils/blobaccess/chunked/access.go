package chunked

import (
	"fmt"
	"io"
	"sync"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/blobaccess"
	"ocm.software/ocm/api/utils/blobaccess/bpi"
	"ocm.software/ocm/api/utils/mime"
)

func newChunck(r io.Reader, fss ...vfs.FileSystem) (bpi.BlobAccess, error) {
	t, err := blobaccess.NewTempFile("", "chunk-*", fss...)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(t.Writer(), r)
	if err != nil {
		t.Close()
		return nil, err
	}
	return t.AsBlob(mime.MIME_OCTET), nil
}

type ChunkedBlobSource interface {
	Next() (bpi.BlobAccess, error)
}

type chunkedAccess struct {
	lock      sync.Mutex
	chunksize int64
	reader    *accessio.ChunkedReader
	fs        vfs.FileSystem
	cont      bool
}

func New(r io.Reader, chunksize int64, fss ...vfs.FileSystem) ChunkedBlobSource {
	reader := accessio.NewChunkedReader(r, chunksize)
	return &chunkedAccess{
		chunksize: chunksize,
		reader:    reader,
		fs:        utils.FileSystem(fss...),
		cont:      false,
	}
}

func (r *chunkedAccess) Next() (bpi.BlobAccess, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.cont {
		if !r.reader.ChunkDone() {
			return nil, fmt.Errorf("unexpected incomplete read")
		}
		if !r.reader.Next() {
			return nil, nil
		}
	}
	r.cont = true
	return newChunck(r.reader, r.fs)
}
