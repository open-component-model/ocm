package blobaccess

import (
	"ocm.software/ocm/api/utils/blobaccess/internal"
)

const (
	KIND_BLOB      = internal.KIND_BLOB
	KIND_MEDIATYPE = internal.KIND_MEDIATYPE

	BLOB_UNKNOWN_SIZE   = internal.BLOB_UNKNOWN_SIZE
	BLOB_UNKNOWN_DIGEST = internal.BLOB_UNKNOWN_DIGEST
)

type (
	DataAccess = internal.DataAccess
	DataReader = internal.DataReader
	DataGetter = internal.DataGetter
)

type (
	BlobAccess         = internal.BlobAccess
	BlobAccessProvider = internal.BlobAccessProvider

	DataSource   = internal.DataSource
	DigestSource = internal.DigestSource
	MimeType     = internal.MimeType
)

type FileLocation = internal.FileLocation
