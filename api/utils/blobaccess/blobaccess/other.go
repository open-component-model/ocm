// Package blobutils provides some utility types and functions
// for blobaccesses, which cannot be put into the blobaccess package,
// because this would introduce cycles in some blobaccess implementation
// packages.
package blobaccess

import (
	"github.com/opencontainers/go-digest"

	"ocm.software/ocm/api/utils/blobaccess/bpi"
	"ocm.software/ocm/api/utils/runtimefinalizer"
)

////////////////////////////////////////////////////////////////////////////////

type blobprovider struct {
	blob bpi.BlobAccess
}

var _ bpi.BlobAccessProvider = (*blobprovider)(nil)

func (b *blobprovider) BlobAccess() (bpi.BlobAccess, error) {
	return b.blob.Dup()
}

func (b *blobprovider) Close() error {
	return b.blob.Close()
}

// ProviderForBlobAccess provides subsequent blob accesses
// as long as the given blob access is not closed.
// If required the blob can be closed with the additionally
// provided Close method.
// ATTENTION: the underlying BlobAccess will not be closed
// as long as the provider is not closed, but the BlobProvider
// interface is no io.Closer.
// To be on the safe side, this method should only be called
// with static blob access, featuring a NOP closer without
// any attached external resources, which should be released.
func ProviderForBlobAccess(blob bpi.BlobAccess) *blobprovider {
	return &blobprovider{blob}
}

type _blobAccess = bpi.BlobAccess

////////////////////////////////////////////////////////////////////////////////

// AnnotatedBlobAccess provides access to the original underlying data source.
type AnnotatedBlobAccess[T bpi.DataAccess] interface {
	_blobAccess
	Source() T
}

type annotatedBlobAccessView[T bpi.DataAccess] struct {
	_blobAccess
	id         runtimefinalizer.ObjectIdentity
	annotation T
}

func (a *annotatedBlobAccessView[T]) Close() error {
	return a._blobAccess.Close()
}

func (a *annotatedBlobAccessView[T]) Dup() (bpi.BlobAccess, error) {
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
// It adds the additional blob access metadata (mime, digest, and size).
// Digest and size can be set to unknown using the constants (BLOB_UNKNOWN_DIGEST
// and BLOB_UNKNOWN_SIZE).
func ForDataAccess[T bpi.DataAccess](digest digest.Digest, size int64, mimeType string, access T) AnnotatedBlobAccess[T] {
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
func WithMimeType(mimeType string, blob bpi.BlobAccess) bpi.BlobAccess {
	return &mimeBlob{blob, mimeType}
}

func (b *mimeBlob) Dup() (bpi.BlobAccess, error) {
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

func NonClosable(blob bpi.BlobAccess) bpi.BlobAccess {
	return &blobNopCloser{blob}
}

func (b *blobNopCloser) Close() error {
	return nil
}
