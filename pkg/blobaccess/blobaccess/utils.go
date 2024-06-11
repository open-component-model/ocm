package blobaccess

import (
	"io"

	"github.com/open-component-model/ocm/pkg/blobaccess/bpi"
	"github.com/open-component-model/ocm/pkg/iotools"
)

func Cast[I interface{}](acc BlobAccess) I {
	return bpi.Cast[I](acc)
}

////////////////////////////////////////////////////////////////////////////////

// BlobData can be applied directly to a function result
// providing a BlobAccess to get the data for the provided blob.
// If the blob access providing function provides an error
// result it is passed to the caller.
func BlobData(blob DataGetter, err ...error) ([]byte, error) {
	if len(err) > 0 && err[0] != nil {
		return nil, err[0]
	}
	return blob.Get()
}

// BlobReader can be applied directly to a function result
// providing a BlobAccess to get a reader for the provided blob.
// If the blob access providing function provides an error
// result it is passed to the caller.
func BlobReader(blob DataReader, err ...error) (io.ReadCloser, error) {
	if len(err) > 0 && err[0] != nil {
		return nil, err[0]
	}
	return blob.Reader()
}

// DataFromProvider extracts the data for a given BlobAccess provider.
func DataFromProvider(s BlobAccessProvider) ([]byte, error) {
	blob, err := s.BlobAccess()
	if err != nil {
		return nil, err
	}
	defer blob.Close()
	return blob.Get()
}

// ReaderFromProvider gets a reader for a BlobAccess provided by
// a BlobAccesssProvider. Closing the Reader also closes the BlobAccess.
func ReaderFromProvider(s BlobAccessProvider) (io.ReadCloser, error) {
	blob, err := s.BlobAccess()
	if err != nil {
		return nil, err
	}
	r, err := blob.Reader()
	if err != nil {
		blob.Close()
		return nil, err
	}
	return iotools.AddReaderCloser(r, blob), nil
}

// MimeReaderFromProvider gets a reader for a BlobAccess provided by
// a BlobAccesssProvider. Closing the Reader also closes the BlobAccess.
// Additionally, the mime type of the blob is returned.
func MimeReaderFromProvider(s BlobAccessProvider) (io.ReadCloser, string, error) {
	blob, err := s.BlobAccess()
	if err != nil {
		return nil, "", err
	}
	mime := blob.MimeType()
	r, err := blob.Reader()
	if err != nil {
		blob.Close()
		return nil, "", err
	}
	return iotools.AddReaderCloser(r, blob), mime, nil
}
