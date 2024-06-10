package blobaccess

import (
	"github.com/mandelsoft/goutils/errors"
	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/blobaccess/bpi"
	"github.com/open-component-model/ocm/pkg/refmgmt"
	"github.com/open-component-model/ocm/pkg/runtimefinalizer"
	"github.com/open-component-model/ocm/pkg/utils"
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

////////////////////////////////////////////////////////////////////////////////

type blobprovider struct {
	blob BlobAccess
}

var _ BlobAccessProvider = (*blobprovider)(nil)

func (b *blobprovider) BlobAccess() (BlobAccess, error) {
	return b.blob.Dup()
}

func (b *blobprovider) Close() error {
	return b.blob.Close()
}

// ProviderForBlobAccess provides subsequent bloc accesses
// as long as the given blob access is not closed.
// If required the blob can be closed with the additionally
// provided Close method.
// ATTENTION: the underlying BlobAccess wil not be closed
// as long as the provider is not closed, but the BlobProvider
// interface is no io.Closer.
// To be on the safe side, this method should only be called
// with static blob access, featuring a NOP closer without
// anny attached external resources, which should be released.
func ProviderForBlobAccess(blob BlobAccess) *blobprovider {
	return &blobprovider{blob}
}

type _blobAccess = BlobAccess

////////////////////////////////////////////////////////////////////////////////

// AnnotatedBlobAccess provides access to the original underlying data source.
type AnnotatedBlobAccess[T DataAccess] interface {
	_blobAccess
	Source() T
}

type annotatedBlobAccessView[T DataAccess] struct {
	_blobAccess
	id         runtimefinalizer.ObjectIdentity
	annotation T
}

func (a *annotatedBlobAccessView[T]) Close() error {
	return a._blobAccess.Close()
}

func (a *annotatedBlobAccessView[T]) Dup() (BlobAccess, error) {
	b, err := a._blobAccess.Dup()
	if err != nil {
		return nil, err
	}
	return &annotatedBlobAccessView[T]{
		id:          runtimefinalizer.NewObjectIdentity(a.id.String()),
		_blobAccess: b,
		annotation:  a.annotation,
	}, nil
}

func (a *annotatedBlobAccessView[T]) Source() T {
	return a.annotation
}

// ForDataAccess wraps the general access object into a blob access.
// It closes the wrapped access, if closed.
// If the wrapped data access does not need a close, the BlobAccess
// does not need a close, also.
func ForDataAccess[T DataAccess](digest digest.Digest, size int64, mimeType string, access T) AnnotatedBlobAccess[T] {
	a := bpi.BaseAccessForDataAccessAndMeta(mimeType, access, digest, size)

	return &annotatedBlobAccessView[T]{
		id:          runtimefinalizer.NewObjectIdentity("annotatedBlobAccess"),
		_blobAccess: bpi.NewBlobAccessForBase(a),
		annotation:  access,
	}
}

////////////////////////////////////////////////////////////////////////////////

type mimeBlob struct {
	_blobAccess
	mimetype string
}

// WithMimeType changes the mime type for a blob access
// by wrapping the given blob access. It does NOT provide
// a new view for the given blob access, so closing the resulting
// blob access will directly close the backing blob access.
func WithMimeType(mimeType string, blob BlobAccess) BlobAccess {
	return &mimeBlob{blob, mimeType}
}

func (b *mimeBlob) Dup() (BlobAccess, error) {
	n, err := b._blobAccess.Dup()
	if err != nil {
		return nil, err
	}
	return &mimeBlob{n, b.mimetype}, nil
}

func (b *mimeBlob) MimeType() string {
	return b.mimetype
}

////////////////////////////////////////////////////////////////////////////////

type blobNopCloser struct {
	_blobAccess
}

func NonClosable(blob BlobAccess) BlobAccess {
	return &blobNopCloser{blob}
}

func (b *blobNopCloser) Close() error {
	return nil
}
