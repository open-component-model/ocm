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
	"compress/bzip2"
	"fmt"
	"io"
)

// Bzip2AlgorithmName is the name used by pkg/compression.Bzip2.
// NOTE: Importing only this /types package does not inherently guarantee a Bzip2 algorithm
// will actually be available. (In fact it is intended for this types package not to depend
// on any of the implementations.)
const Bzip2AlgorithmName = "bzip2"

// Bzip2 compression.
var Bzip2 = NewAlgorithm(Bzip2AlgorithmName, Bzip2AlgorithmName,
	[]byte{0x42, 0x5A, 0x68}, bzip2Decompressor, bzip2Compressor)

func init() {
	Register(Bzip2)
}

// bzip2Decompressor is a DecompressorFunc for the bzip2 compression algorithm.
func bzip2Decompressor(r io.Reader) (io.ReadCloser, error) {
	return io.NopCloser(bzip2.NewReader(r)), nil
}

// bzip2Compressor is a CompressorFunc for the bzip2 compression algorithm.
func bzip2Compressor(r io.Writer, metadata map[string]string, level *int) (io.WriteCloser, error) {
	return nil, fmt.Errorf("bzip2 compression not supported")
}
