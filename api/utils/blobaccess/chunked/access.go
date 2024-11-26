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

func newChunk(r io.Reader, fss ...vfs.FileSystem) (bpi.BlobAccess, error) {
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

// ChunkedBlobSource provides a sequence of
// bpi.BlobAccess objects.
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

// New provides a sequence of
// bpi.BlobAccess objects for a given io.Reader
// each with a limited size.
// The provided blobs are temporarily stored
// on the filesystem and can therefore be kept
// and accessed any number of times until they are closed.
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
	return newChunk(r.reader, r.fs)
}
