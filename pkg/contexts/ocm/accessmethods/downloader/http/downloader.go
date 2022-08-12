package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

// Downloader simply uses the default HTTP client to download the contents of a URL.
type Downloader struct {
	link string
}

func NewDownloader(link string) *Downloader {
	return &Downloader{
		link: link,
	}
}

func (h *Downloader) Download(w io.WriterAt) error {
	resp, err := http.Get(h.link)
	if err != nil {
		return fmt.Errorf("failed to get link: %w", err)
	}
	defer resp.Body.Close()
	var blob []byte
	buf := bytes.NewBuffer(blob)
	if _, err := io.Copy(buf, resp.Body); err != nil {
		return fmt.Errorf("failed to copy response body: %w", err)
	}
	if _, err := w.WriteAt(buf.Bytes(), 0); err != nil {
		return fmt.Errorf("failed to WriteAt to the writer: %w", err)
	}
	return nil
}
