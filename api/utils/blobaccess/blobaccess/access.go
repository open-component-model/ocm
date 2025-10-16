package blobaccess

import (
	"bytes"
	"io"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/opencontainers/go-digest"
	"ocm.software/ocm/api/utils/blobaccess/bpi"
	"ocm.software/ocm/api/utils/iotools"
	mimetypes "ocm.software/ocm/api/utils/mime"
)

type bytesAccess struct {
	_nopCloser
	data   []byte
	origin string
}

func DataAccessForData(data []byte, origin ...string) bpi.DataSource {
	path := ""
	if len(origin) > 0 {
		path = filepath.Join(origin...)
	}
	return &bytesAccess{data: data, origin: path}
}

func DataAccessForString(data string, origin ...string) bpi.DataSource {
	return DataAccessForData([]byte(data), origin...)
}

func (a *bytesAccess) Get() ([]byte, error) {
	return a.data, nil
}

func (a *bytesAccess) Reader() (io.ReadCloser, error) {
	return iotools.ReadCloser(bytes.NewReader(a.data)), nil
}

func (a *bytesAccess) Origin() string {
	return a.origin
}

////////////////////////////////////////////////////////////////////////////////

// ForString wraps a string into a BlobAccess, which does not need a close.
func ForString(mime string, data string) bpi.BlobAccess {
	if mime == "" {
		mime = mimetypes.MIME_TEXT
	}
	return ForData(mime, []byte(data))
}

func ProviderForString(mime, data string) bpi.BlobAccessProvider {
	return bpi.BlobAccessProviderFunction(func() (bpi.BlobAccess, error) {
		return ForString(mime, data), nil
	})
}

// ForData wraps data into a BlobAccess, which does not need a close.
func ForData(mime string, data []byte) bpi.BlobAccess {
	if mime == "" {
		mime = mimetypes.MIME_OCTET
	}
	return bpi.ForStaticDataAccessAndMeta(mime, DataAccessForData(data), digest.FromBytes(data), int64(len(data)))
}

func ProviderForData(mime string, data []byte) bpi.BlobAccessProvider {
	return bpi.BlobAccessProviderFunction(func() (bpi.BlobAccess, error) {
		return ForData(mime, data), nil
	})
}
