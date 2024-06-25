package iotools

import (
	"crypto"
	"fmt"
	"hash"
	"io"
	"net/http"
	"strings"

	"github.com/mandelsoft/goutils/errors"
)

type Hashes map[crypto.Hash]hash.Hash

func NewHashes(algorithms ...crypto.Hash) Hashes {
	hashMap := make(Hashes, len(algorithms))
	for _, algorithm := range algorithms {
		hashMap[algorithm] = algorithm.New()
	}
	return hashMap
}

func (h Hashes) Write(c int, buf []byte) {
	if c > 0 {
		for _, hash := range h {
			hash.Write(buf[:c])
		}
	}
}

func (h Hashes) AsHttpHeader() http.Header {
	headers := make(http.Header, len(h))
	for algorithm := range h {
		headers.Set(headerName(algorithm), h.GetString(algorithm))
	}
	return headers
}

func (h Hashes) GetBytes(algorithm crypto.Hash) []byte {
	hash := h[algorithm]
	if hash != nil {
		return hash.Sum(nil)
	}
	return nil
}

func (h Hashes) GetString(algorithm crypto.Hash) string {
	return fmt.Sprintf("%x", h.GetBytes(algorithm))
}

func headerName(algorithm crypto.Hash) string {
	a := strings.ReplaceAll(algorithm.String(), "-", "")
	return "X-Checksum-" + a[:1] + strings.ToLower(a[1:])
}

////////////////////////////////////////////////////////////////////////////////

type HashReader struct {
	reader  io.Reader
	hashMap Hashes
}

func NewHashReader(delegate io.Reader, algorithms ...crypto.Hash) *HashReader {
	newInstance := HashReader{
		reader:  delegate,
		hashMap: NewHashes(algorithms...),
	}
	return &newInstance
}

func (h *HashReader) Read(buf []byte) (int, error) {
	c, err := h.reader.Read(buf)
	if err == nil {
		h.hashMap.Write(c, buf)
	}
	return c, err
}

func (h *HashReader) Hashes() Hashes {
	return h.hashMap
}

func (h *HashReader) ReadAll() ([]byte, error) {
	return io.ReadAll(h.reader)
}

// CalcHashes returns the total number of bytes read and an error if any besides EOF.
func (h *HashReader) CalcHashes() (int64, error) {
	b := make([]byte, 0, 512)
	cnt := int64(0)
	for {
		n, err := h.Read(b[0:cap(b)]) // read a chunk, always from the beginning
		b = b[:n]                     // reset slice to the actual read bytes
		cnt += int64(n)
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
			}
			return cnt, err
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

type HashWriter struct {
	writer  io.Writer
	hashMap Hashes
}

func NewHashWriter(w io.Writer, algorithms ...crypto.Hash) *HashWriter {
	newInstance := HashWriter{
		writer:  w,
		hashMap: NewHashes(algorithms...),
	}
	return &newInstance
}

func (h *HashWriter) Write(buf []byte) (int, error) {
	c, err := h.writer.Write(buf)
	if err == nil {
		h.hashMap.Write(c, buf)
	}
	return c, err
}

func (h *HashWriter) Hashes() Hashes {
	return h.hashMap
}
