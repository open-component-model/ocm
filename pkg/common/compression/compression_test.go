package compression

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchReader(t *testing.T) {
	content := "this is my content"
	cases := []string{"this", "th", "this is"}
	r := bytes.NewBuffer([]byte(content))
	mr := NewMatchReader(r)
	for _, ca := range cases {
		mr.Reset()
		buf := [20]byte{}
		n, err := io.ReadAtLeast(mr, buf[:len(ca)], len(ca))
		require.NoError(t, err, ca)
		assert.Equal(t, []byte(ca), buf[:n], ca)

	}
	reader := mr.Reader()
	all, err := io.ReadAll(reader)
	require.NoError(t, err, "all")
	assert.Equal(t, []byte(content), all, "all")
}

func TestDetectCompression(t *testing.T) {
	cases := []string{
		"fixtures/Hello.uncompressed",
		"fixtures/Hello.gz",
		"fixtures/Hello.bz2",
		"fixtures/Hello.xz",
		"fixtures/Hello.zst",
	}

	// The original stream is preserved.
	fmt.Printf("***preserve***\n")
	for _, c := range cases {
		originalContents, err := ioutil.ReadFile(c)
		require.NoError(t, err, c)

		stream, err := os.Open(c)
		require.NoError(t, err, c)
		defer stream.Close()

		fmt.Printf("preserve %s\n", c)
		_, updatedStream, err := DetectCompression(stream)
		require.NoError(t, err, c)

		updatedContents, err := ioutil.ReadAll(updatedStream)
		require.NoError(t, err, c)
		assert.Equal(t, originalContents, updatedContents, c)
	}

	// The correct decompressor is chosen, and the result is as expected.
	fmt.Printf("***decompress***\n")
	for _, c := range cases {
		stream, err := os.Open(c)
		require.NoError(t, err, c)
		defer stream.Close()

		algo, updatedStream, err := DetectCompression(stream)
		require.NoError(t, err, c)

		fmt.Printf("decompress %s -> %s\n", c, algo.Name())

		s, err := algo.Decompressor(updatedStream)
		require.NoError(t, err)
		defer s.Close()
		updatedStream = s

		uncompressedContents, err := ioutil.ReadAll(updatedStream)
		require.NoError(t, err, c)
		assert.Equal(t, []byte("Hello"), uncompressedContents, c)
	}

	fmt.Printf("***empty***\n")

	// Empty input is handled reasonably.
	algo, updatedStream, err := DetectCompression(bytes.NewReader([]byte{}))
	require.NoError(t, err)
	assert.Equal(t, None, algo)
	updatedContents, err := ioutil.ReadAll(updatedStream)
	require.NoError(t, err)
	assert.Equal(t, []byte{}, updatedContents)

	// Error reading input
	reader, writer := io.Pipe()
	defer reader.Close()
	err = writer.CloseWithError(errors.New("Expected error reading input in DetectCompression"))
	assert.NoError(t, err)
	_, _, err = DetectCompression(reader)
	assert.Error(t, err)
}

func TestAutoDecompress(t *testing.T) {
	cases := []struct {
		filename     string
		isCompressed bool
	}{
		{"fixtures/Hello.uncompressed", false},
		{"fixtures/Hello.gz", true},
		{"fixtures/Hello.bz2", true},
		{"fixtures/Hello.xz", true},
	}

	// The correct decompressor is chosen, and the result is as expected.
	for _, c := range cases {
		stream, err := os.Open(c.filename)
		require.NoError(t, err, c.filename)
		defer stream.Close()

		uncompressedStream, isCompressed, err := AutoDecompress(stream)
		require.NoError(t, err, c.filename)
		defer uncompressedStream.Close()

		assert.Equal(t, c.isCompressed, isCompressed)

		uncompressedContents, err := ioutil.ReadAll(uncompressedStream)
		require.NoError(t, err, c.filename)
		assert.Equal(t, []byte("Hello"), uncompressedContents, c.filename)
	}

	// Empty input is handled reasonably.
	uncompressedStream, isCompressed, err := AutoDecompress(bytes.NewReader([]byte{}))
	require.NoError(t, err)
	assert.False(t, isCompressed)
	uncompressedContents, err := ioutil.ReadAll(uncompressedStream)
	require.NoError(t, err)
	assert.Equal(t, []byte{}, uncompressedContents)

	// Error initializing a decompressor (for a detected format)
	_, _, err = AutoDecompress(bytes.NewReader([]byte{0xFD, 0x37, 0x7A, 0x58, 0x5A, 0x00}))
	assert.Error(t, err)

	// Error reading input
	reader, writer := io.Pipe()
	defer reader.Close()
	err = writer.CloseWithError(errors.New("Expected error reading input in AutoDecompress"))
	require.NoError(t, err)
	_, _, err = AutoDecompress(reader)
	assert.Error(t, err)
}
