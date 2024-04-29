package iotools

import (
	"crypto"
	"fmt"
	"hash"
	"io"
	"strings"

	"github.com/open-component-model/ocm/pkg/errors"
)

type HashReader struct {
	reader  io.Reader
	hashMap map[crypto.Hash]hash.Hash
}

func NewHashReader(delegate io.Reader, algorithms ...crypto.Hash) *HashReader {
	newInstance := HashReader{
		reader:  delegate,
		hashMap: initMap(algorithms),
	}
	return &newInstance
}

func (h *HashReader) Read(buf []byte) (int, error) {
	c, err := h.reader.Read(buf)
	return write(h, c, buf, err)
}

func (h *HashReader) GetString(algorithm crypto.Hash) string {
	return getString(h, algorithm)
}

func (h *HashReader) GetBytes(algorithm crypto.Hash) []byte {
	return getBytes(h, algorithm)
}

func (h *HashReader) HttpHeader() map[string]string {
	return httpHeader(h)
}

func (h *HashReader) hashes() map[crypto.Hash]hash.Hash {
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
	hashMap map[crypto.Hash]hash.Hash
}

func NewHashWriter(w io.Writer, algorithms ...crypto.Hash) *HashWriter {
	newInstance := HashWriter{
		writer:  w,
		hashMap: initMap(algorithms),
	}
	return &newInstance
}

func (h *HashWriter) Write(buf []byte) (int, error) {
	c, err := h.writer.Write(buf)
	return write(h, c, buf, err)
}

func (h *HashWriter) GetString(algorithm crypto.Hash) string {
	return getString(h, algorithm)
}

func (h *HashWriter) GetBytes(algorithm crypto.Hash) []byte {
	return getBytes(h, algorithm)
}

func (h *HashWriter) HttpHeader() map[string]string {
	return httpHeader(h)
}

func (h *HashWriter) hashes() map[crypto.Hash]hash.Hash {
	return h.hashMap
}

////////////////////////////////////////////////////////////////////////////////

type hashes interface {
	hashes() map[crypto.Hash]hash.Hash
}

func getString(h hashes, algorithm crypto.Hash) string {
	return fmt.Sprintf("%x", getBytes(h, algorithm))
}

func getBytes(h hashes, algorithm crypto.Hash) []byte {
	hash := h.hashes()[algorithm]
	if hash != nil {
		return hash.Sum(nil)
	}
	return nil
}

func httpHeader(h hashes) map[string]string {
	headers := make(map[string]string, len(h.hashes()))
	for algorithm := range h.hashes() {
		headers[headerName(algorithm)] = getString(h, algorithm)
	}
	return headers
}

func initMap(algorithms []crypto.Hash) map[crypto.Hash]hash.Hash {
	hashMap := make(map[crypto.Hash]hash.Hash, len(algorithms))
	for _, algorithm := range algorithms {
		hashMap[algorithm] = algorithm.New()
	}
	return hashMap
}

func write(h hashes, c int, buf []byte, err error) (int, error) {
	if err == nil && c > 0 {
		for _, hash := range h.hashes() {
			hash.Write(buf[:c])
		}
	}
	return c, err
}

////////////////////////////////////////////////////////////////////////////////

func headerName(hash crypto.Hash) string {
	a := strings.ReplaceAll(hash.String(), "-", "")
	return "X-Checksum-" + a[:1] + strings.ToLower(a[1:])
}
