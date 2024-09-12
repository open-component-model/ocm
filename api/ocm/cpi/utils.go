package cpi

import (
	"io"

	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
)

type AccessMethodSource interface {
	AccessMethod() (AccessMethod, error)
}

// GetResourceMimeReader gets a Reader for a given resource/source access.
// It provides a Reader handling the Close contract for the access method.
func GetResourceReader(acc AccessProvider) (io.ReadCloser, error) {
	return blobaccess.ReaderFromProvider(acc)
}

// GetResourceMimeReader gets a Reader for a given resource/source access.
// It provides a Reader handling the Close contract for the access method.
// Additionally, the mime type is returned.
func GetResourceMimeReader(acc AccessProvider) (io.ReadCloser, string, error) {
	return blobaccess.MimeReaderFromProvider(acc)
}

////////////////////////////////////////////////////////////////////////////////

func ArtifactNameHint(spec AccessSpec, cv ComponentVersionAccess) string {
	if h, ok := spec.(HintProvider); ok {
		return h.GetReferenceHint(cv)
	}
	return ""
}

func ReferenceHint(spec AccessSpec, cv ComponentVersionAccess) string {
	if h, ok := spec.(HintProvider); ok {
		return h.GetReferenceHint(cv)
	}
	return ""
}

func GlobalAccess(spec AccessSpec, ctx Context) AccessSpec {
	g := spec.GlobalAccessSpec(ctx)
	if g != nil && g.IsLocal(ctx) {
		g = nil
	}
	return g
}
