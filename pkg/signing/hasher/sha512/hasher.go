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

package sha512

import (
	"crypto"
	"crypto/sha512"
	"hash"

	"github.com/open-component-model/ocm/pkg/signing"
)

const Algorithm = "sha512"

func init() {
	signing.DefaultHandlerRegistry().RegisterHasher(Handler{})
}

// Handler is a signatures.Hasher compatible struct to hash with sha256.
type Handler struct{}

var _ signing.Hasher = Handler{}

func (_ Handler) Algorithm() string {
	return Algorithm
}

// Create creates a Hasher instance for no digest.
func (_ Handler) Create() hash.Hash {
	return sha512.New()
}

func (_ Handler) Crypto() crypto.Hash {
	return crypto.SHA512
}
