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

package signing

import (
	"crypto"
	"encoding/json"
	"hash"
)

type Signature struct {
	Value     string
	MediaType string
	Algorithm string
	Issuer    string
}

func (s *Signature) String() string {
	data, _ := json.Marshal(s)
	return string(data)
}

// Signer interface is used to implement different signing algorithms.
// Each Signer should have a matching Verifier.
type Signer interface {
	// Sign returns the signature for the given digest
	Sign(digest string, hash crypto.Hash, issuer string, privatekey interface{}) (*Signature, error)
	// Algorithm is the name of the finally used signature algorithm.
	// A signer might be registered using a logical name, so there might
	// be multiple signer registration providing the same signature algorithm
	Algorithm() string
}

// Verifier interface is used to implement different verification algorithms.
// Each Verifier should have a matching Signer.
type Verifier interface {
	// Verify checks the signature, returns an error on verification failure
	Verify(digest string, hash crypto.Hash, sig *Signature, publickey interface{}) error
	Algorithm() string
}

// SignatureHandler can create and verify signature of a dedicated type
type SignatureHandler interface {
	Algorithm() string
	Signer
	Verifier
}

// Hasher creates a new hash.Hash interface
type Hasher interface {
	Algorithm() string
	Create() hash.Hash
	Crypto() crypto.Hash
}
