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

// This package has been initially taken from github.com/containers/image
// and modified to be provide a useful simple API based on
// an Algorithm interface

package compression

import (
	"io"
)

// CompressorFunc writes the compressed stream to the given writer using the specified compression level.
// The caller must call Close() on the stream (even if the input stream does not need closing!).
type CompressorFunc func(io.Writer, map[string]string, *int) (io.WriteCloser, error)

// DecompressorFunc returns the decompressed stream, given a compressed stream.
// The caller must call Close() on the decompressed stream (even if the compressed input stream does not need closing!).
type DecompressorFunc func(io.Reader) (io.ReadCloser, error)

// Algorithm is a compression algorithm provided and supported by pkg/compression.
// It can’t be supplied from the outside.
type Algorithm interface {
	Name() string
	Compressor(io.Writer, map[string]string, *int) (io.WriteCloser, error)
	Decompressor(io.Reader) (io.ReadCloser, error)
	Match(MatchReader) (bool, error)
}
