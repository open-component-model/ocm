package iotools

import (
	"io"
)

type NopCloser struct{}

func (NopCloser) Close() error {
	return nil
}

type NopWriter struct {
	NopCloser
}

var _ io.Writer = NopWriter{}

func (n2 NopWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
