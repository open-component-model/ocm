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
	"io"

	"github.com/klauspost/pgzip"
)

// GzipAlgorithmName is the name used by pkg/compression.Gzip.
// NOTE: Importing only this /types package does not inherently guarantee a Gzip algorithm
// will actually be available. (In fact it is intended for this types package not to depend
// on any of the implementations.)
const GzipAlgorithmName = "gzip"

var Gzip = NewAlgorithm(GzipAlgorithmName, GzipAlgorithmName,
	[]byte{0x1F, 0x8B, 0x08}, gzipDecompressor, gzipCompressor)

func init() {
	Register(Gzip)
}

// gzipDecompressor is a DecompressorFunc for the gzip compression algorithm.
func gzipDecompressor(r io.Reader) (io.ReadCloser, error) {
	return pgzip.NewReader(r)
}

// gzipCompressor is a CompressorFunc for the gzip compression algorithm.
func gzipCompressor(r io.Writer, metadata map[string]string, level *int) (io.WriteCloser, error) {
	if level != nil {
		return pgzip.NewWriterLevel(r, *level)
	}
	return pgzip.NewWriter(r), nil
}
