package tarutils

import (
	"compress/gzip"
	"io"
)

func Gzip(w io.Writer) io.WriteCloser {
	return gzip.NewWriter(w)
}
