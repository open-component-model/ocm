package bpi

import (
	"github.com/mandelsoft/goutils/errors"
	"github.com/opencontainers/go-digest"

	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/blobaccess/internal"
	"ocm.software/ocm/api/utils/refmgmt"
)

const (
	KIND_BLOB      = internal.KIND_BLOB
	KIND_MEDIATYPE = internal.KIND_MEDIATYPE

	BLOB_UNKNOWN_SIZE   = internal.BLOB_UNKNOWN_SIZE
	BLOB_UNKNOWN_DIGEST = internal.BLOB_UNKNOWN_DIGEST
)

var ErrClosed = refmgmt.ErrClosed

type DataAccess = internal.DataAccess

type (
	BlobAccess         = internal.BlobAccess
	BlobAccessBase     = internal.BlobAccessBase
	BlobAccessProvider = internal.BlobAccessProvider

	Validatable = utils.Validatable

	DataReader   = internal.DataReader
	DataGetter   = internal.DataGetter
	DataSource   = internal.DataSource
	DigestSource = internal.DigestSource
	MimeType     = internal.MimeType
)

type FileLocation = internal.FileLocation

type BlobAccessProviderFunction func() (BlobAccess, error)

func (p BlobAccessProviderFunction) BlobAccess() (BlobAccess, error) {
	return p()
}

func ErrBlobNotFound(digest digest.Digest) error {
	return errors.ErrNotFound(KIND_BLOB, digest.String())
}

func IsErrBlobNotFound(err error) bool {
	return errors.IsErrNotFoundKind(err, KIND_BLOB)
}
