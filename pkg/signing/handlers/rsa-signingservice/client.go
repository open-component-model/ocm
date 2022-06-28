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
	"bytes"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
)

const (
	AcceptHeader   = "Accept"
	AlgorithHeader = "X-SignatatureAlgorithm"

	// MediaTypePEM defines the media type for PEM formatted data.
	MediaTypePEM = "application/x-pem-file"
)

type SigningServerSigner struct {
	Url      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewSigningClient(url string, username, password string) (*SigningServerSigner, error) {
	return &SigningServerSigner{
		url, username, password,
	}, nil
}

func (signer *SigningServerSigner) Sign(algo string, digest string, issuer string, key interface{}) (*signing.Signature, error) {
	decodedHash, err := hex.DecodeString(digest)
	if err != nil {
		return nil, fmt.Errorf("failed decoding hash: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/sign", signer.Url), bytes.NewBuffer(decodedHash))
	if err != nil {
		return nil, fmt.Errorf("failed building http request: %w", err)
	}
	req.Header.Add(AcceptHeader, MediaTypePEM)
	req.Header.Add(SignaturePEMBlockAlgorithmHeader, algo)
	req.SetBasicAuth(signer.Username, signer.Password)

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed sending request: %w", err)
	}
	defer res.Body.Close()

	responseBodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request returned with response code %d: %s", res.StatusCode, string(responseBodyBytes))
	}

	signaturePemBlocks, err := rsa.GetSignaturePEMBlocks(responseBodyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed getting signature pem block from response: %w", err)
	}

	if len(signaturePemBlocks) != 1 {
		return nil, fmt.Errorf("expected 1 signature pem block, found %d", len(signaturePemBlocks))
	}
	signatureBlock := signaturePemBlocks[0]

	signature := signatureBlock.Bytes
	if len(signature) == 0 {
		return nil, errors.New("invalid response: signature block doesn't contain signature")
	}

	algorithm := signatureBlock.Headers[rsa.SignaturePEMBlockAlgorithmHeader]
	if algorithm == "" {
		return nil, fmt.Errorf("invalid response: %s header is empty", rsa.SignaturePEMBlockAlgorithmHeader)
	}

	encodedSignature := pem.EncodeToMemory(signatureBlock)

	return &signing.Signature{
		Value:     string(encodedSignature),
		MediaType: MediaTypePEM,
		Algorithm: algorithm,
		Issuer:    issuer,
	}, nil
}
