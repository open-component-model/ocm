package iotools

import (
	"io"

	"github.com/mandelsoft/goutils/sliceutils"
)

// Deprecated: use AddReaderCloser .
func AddCloser(reader io.ReadCloser, closer io.Closer, msg ...string) io.ReadCloser {
	return AddReaderCloser(reader, closer, sliceutils.Convert[any](msg)...)
}
