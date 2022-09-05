// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package compression

import (
	"bytes"
	"errors"
	"io"
)

// algorithm is a default implementation for Algorithm that can be used for CompressStream
// based on Compression and DEcompression functions.
type algorithm struct {
	name         string
	mime         string
	prefix       []byte // Initial bytes of a stream compressed using this algorithm, or empty to disable detection.
	decompressor DecompressorFunc
	compressor   CompressorFunc
}

// NewAlgorithm creates an Algorithm instance.
// This function exists so that Algorithm instances can only be created by code that
// is allowed to import this internal subpackage.
func NewAlgorithm(name, mime string, prefix []byte, decompressor DecompressorFunc, compressor CompressorFunc) Algorithm {
	return &algorithm{
		name:         name,
		mime:         mime,
		prefix:       prefix,
		decompressor: decompressor,
		compressor:   compressor,
	}
}

// Name returns the name for the compression algorithm.
func (c *algorithm) Name() string {
	return c.name
}

// InternalUnstableUndocumentedMIMEQuestionMark ???
// DO NOT USE THIS anywhere outside of c/image until it is properly documented.
func (c *algorithm) InternalUnstableUndocumentedMIMEQuestionMark() string {
	return c.mime
}

// Compressor returns a compressor for the given stream according to this algorithm .
func (c *algorithm) Compressor(w io.Writer, meta map[string]string, level *int) (io.WriteCloser, error) {
	if meta == nil {
		meta = map[string]string{}
	}

	return c.compressor(w, meta, level)
}

// Decompressor returns a decompressor for the given stream according to this algorithm .
func (c *algorithm) Decompressor(r io.Reader) (io.ReadCloser, error) {
	return c.decompressor(r)
}

func (c *algorithm) Match(r MatchReader) (bool, error) {
	if len(c.prefix) == 0 {
		return false, nil
	}
	buf := make([]byte, len(c.prefix))
	n, err := io.ReadAtLeast(r, buf, len(buf))
	// fmt.Printf("%s: found %v\n", c.Name(), buf[:n])
	if err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) {
			err = nil
		}
		return false, err
	}
	return bytes.HasPrefix(buf[:n], c.prefix), nil
}
