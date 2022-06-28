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

package rsa_signingservice

import (
	"crypto"
	"fmt"

	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
)

// Algorithm defines the type for the RSA PKCS #1 v1.5 signature algorithm
const Algorithm = rsa.Algorithm
const Name = "rsa-signingsservice"

type Key struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// SignaturePEMBlockAlgorithmHeader defines the header in a signature pem block where the signature algorithm is defined.
const SignaturePEMBlockAlgorithmHeader = "Algorithm"

func init() {
	signing.DefaultHandlerRegistry().RegisterSigner(Name, Handler{})
}

// Handler is a signatures.Signer compatible struct to sign with RSASSA-PKCS1-V1_5.
// using a signature service
type Handler struct {
}

var _ Handler = Handler{}

func (h Handler) Algorithm() string {
	return Algorithm
}

func (h Handler) Sign(digest string, hash crypto.Hash, issuer string, key interface{}) (signature *signing.Signature, err error) {
	privateKey, err := PrivateKey(key)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid rsa private key")
	}
	server, err := NewSigningClient(privateKey.URL, privateKey.Username, privateKey.Password)
	if err != nil {
		return nil, err
	}
	return server.Sign(h.Algorithm(), digest, issuer, key)
}

func PrivateKey(k interface{}) (*Key, error) {
	switch t := k.(type) {
	case *Key:
		return t, nil
	case []byte:
		key := &Key{}
		err := runtime.DefaultYAMLEncoding.Unmarshal(t, key)
		if err != nil {
			return nil, err
		}
		return key, err
	default:
		return nil, fmt.Errorf("unknown key specification %T", k)
	}
}
