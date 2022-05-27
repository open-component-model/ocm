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

package v1

const (
	// ExcludeFromSignature used in digest field for normalisationAlgorithm (in combination with NoDigest for hashAlgorithm and value)
	// to indicate the resource content should not be part of the signature
	ExcludeFromSignature = "EXCLUDE-FROM-SIGNATURE"

	// NoDigest used in digest field for hashAlgorithm and value (in combination with ExcludeFromSignature for normalisationAlgorithm)
	// to indicate the resource content should not be part of the signature
	NoDigest = "NO-DIGEST"
)

// Signatures is a list of signatures
type Signatures []Signature

func (s Signatures) Len() int {
	return len(s)
}

func (s Signatures) Get(i int) *Signature {
	if i >= 0 && i < len(s) {
		return &s[i]
	}
	return nil
}

func (s Signatures) Copy() Signatures {
	if s == nil {
		return nil
	}
	out := make(Signatures, s.Len())
	for i, v := range s {
		out[i] = *v.Copy()
	}
	return out
}

// DigestSpec defines a digest.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type DigestSpec struct {
	HashAlgorithm          string `json:"hashAlgorithm"`
	NormalisationAlgorithm string `json:"normalisationAlgorithm"`
	Value                  string `json:"value"`
}

// Copy provides a copy of the digest spec
func (d *DigestSpec) Copy() *DigestSpec {
	if d == nil {
		return nil
	}
	r := *d
	return &r
}

// SignatureSpec defines a signature.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type SignatureSpec struct {
	Algorithm string `json:"algorithm"`
	Value     string `json:"value"`
	MediaType string `json:"mediaType"`
}

// Signature defines a digest and corresponding signature, identifyable by name.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type Signature struct {
	Name      string        `json:"name"`
	Digest    DigestSpec    `json:"digest"`
	Signature SignatureSpec `json:"signature"`
}

// Copy provides a copy of the signature data
func (s *Signature) Copy() *Signature {
	if s == nil {
		return nil
	}
	r := *s
	return &r
}

//NewExcludeFromSignatureDigest returns the special digest notation to indicate the resource content should not be part of the signature
func NewExcludeFromSignatureDigest() *DigestSpec {
	return &DigestSpec{
		HashAlgorithm:          NoDigest,
		NormalisationAlgorithm: ExcludeFromSignature,
		Value:                  NoDigest,
	}
}
