package blobaccess

import (
	"github.com/mandelsoft/goutils/errors"
	"github.com/opencontainers/go-digest"

	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/refmgmt"
)

var ErrClosed = refmgmt.ErrClosed

func ErrBlobNotFound(digest digest.Digest) error {
	return errors.ErrNotFound(KIND_BLOB, digest.String())
}

func IsErrBlobNotFound(err error) bool {
	return errors.IsErrNotFoundKind(err, KIND_BLOB)
}

////////////////////////////////////////////////////////////////////////////////

// Validatable is an optional interface for DataAccess
// implementations or any other object, which might reach
// an error state. The error can then be queried with
// the method ErrorProvider.Validate.
// This is used to support objects with access methods not
// returning an error. If the object is not valid,
// those methods return an unknown/default state, but
// the object should be queryable for its state.
type Validatable = utils.Validatable

// Validate checks whether a blob access
// is in error state. If yes, an appropriate
// error is returned.
func Validate(o BlobAccess) error {
	return utils.ValidateObject(o)
}
