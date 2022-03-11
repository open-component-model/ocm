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
	"io/ioutil"

	"github.com/ulikunitz/xz"
)

// XzAlgorithmName is the name used by pkg/compression.Xz.
// NOTE: Importing only this /types package does not inherently guarantee a Xz algorithm
// will actually be available. (In fact it is intended for this types package not to depend
// on any of the implementations.)
const XzAlgorithmName = "Xz"

// Xz compression.
var Xz = NewAlgorithm(XzAlgorithmName, XzAlgorithmName,
	[]byte{0xFD, 0x37, 0x7A, 0x58, 0x5A, 0x00}, xzDecompressor, xzCompressor)

func init() {
	Register(Xz)
}

// xzDecompressor is a DecompressorFunc for the xz compression algorithm.
func xzDecompressor(r io.Reader) (io.ReadCloser, error) {
	r, err := xz.NewReader(r)
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(r), nil
}

// xzCompressor is a CompressorFunc for the xz compression algorithm.
func xzCompressor(r io.Writer, metadata map[string]string, level *int) (io.WriteCloser, error) {
	return xz.NewWriter(r)
}
