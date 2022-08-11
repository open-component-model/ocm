package github

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

// Downloader defines an abstraction for downloading an archive from GitHub.
type Downloader interface {
	Download(link string) ([]byte, error)
}

// HTTPDownloader simply uses the default HTTP client to download the contents of a URL.
type HTTPDownloader struct{}

func (h *HTTPDownloader) Download(link string) ([]byte, error) {
	httpResp, err := http.Get(link)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := httpResp.Body.Close(); err != nil {
			fmt.Println("failed to close body: ", err)
		}
	}()

	var blob []byte
	buf := bytes.NewBuffer(blob)
	if _, err := io.Copy(buf, httpResp.Body); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
