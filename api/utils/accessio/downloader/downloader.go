package downloader

import "io"

// Downloader defines a downloader for various objects using a WriterAt to
// transfer data to.
type Downloader interface {
	Download(w io.WriterAt) error
}
