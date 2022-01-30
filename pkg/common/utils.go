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

package common

import (
	"strings"

	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/opencontainers/go-digest"
)

func Digest(access accessio.DataAccess) (digest.Digest, error) {
	reader, err := access.Reader()
	if err != nil {
		return "", err
	}
	defer reader.Close()

	dig, err := digest.FromReader(reader)
	if err != nil {
		return "", err
	}
	return dig, nil
}

// DigestToFileName returns teh file name for a digest
func DigestToFileName(digest digest.Digest) string {
	return strings.Replace(digest.String(), ":", "+", 1)
}

// PathToDigest retuurns the digest encoded into a file name
func PathToDigest(path string) digest.Digest {
	n := filepath.Base(path)
	idx := strings.LastIndex(n, "+")
	if idx < 0 {
		return ""
	}
	return digest.Digest(n[:idx] + ":" + n[idx+1:])
}
