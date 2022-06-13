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

package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"encoding/pem"
	"fmt"

	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/signing"
)

// Algorithm defines the type for the RSA PKCS #1 v1.5 signature algorithm
const Algorithm = "RSASSA-PKCS1-V1_5"

// MediaType defines the media type for a plain RSA signature.
const MediaType = "application/vnd.ocm.signature.rsa"

// MediaTypePEM defines the media type for PEM formatted data.
const MediaTypePEM = "application/x-pem-file"

// SignaturePEMBlockType defines the type of a signature pem block.
const SignaturePEMBlockType = "SIGNATURE"

// SignaturePEMBlockAlgorithmHeader defines the header in a signature pem block where the signature algorithm is defined.
const SignaturePEMBlockAlgorithmHeader = "Algorithm"

func init() {
	signing.DefaultHandlerRegistry().RegisterSigner(Algorithm, Handler{})
}

type PrivateKey = rsa.PrivateKey
type PublicKey = rsa.PublicKey

// Handler is a signatures.Signer compatible struct to sign with RSASSA-PKCS1-V1_5.
// and a signatures.Verifier compatible struct to verify RSASSA-PKCS1-V1_5 signatures.
type Handler struct {
}

var _ Handler = Handler{}

func (h Handler) Algorithm() string {
	return Algorithm
}

func (h Handler) Sign(digest string, key interface{}) (signature *signing.Signature, err error) {
	privateKey, err := GetPrivateKey(key)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid rsa private key")
	}
	decodedHash, err := hex.DecodeString(digest)
	if err != nil {
		return nil, fmt.Errorf("failed decoding hash to bytes")
	}
	// ensure length of hash is correct
	//if len(decodedHash) != 32 {
	//	return "", "", fmt.Errorf("hash to sign has invalid length")
	//}
	sig, err := rsa.SignPKCS1v15(rand.Reader, privateKey, 0, decodedHash)
	if err != nil {
		return nil, fmt.Errorf("failed signing hash, %w", err)
	}
	return &signing.Signature{
		Value:     hex.EncodeToString(sig),
		MediaType: MediaType,
		Algorithm: Algorithm,
	}, nil
}

// Verify checks the signature, returns an error on verification failure
func (h Handler) Verify(digest string, signature string, mediatype string, key interface{}) (err error) {
	var signatureBytes []byte
	publicKey, err := GetPublicKey(key)
	if err != nil {
		return err
	}
	switch mediatype {
	case MediaType:
		signatureBytes, err = hex.DecodeString(signature)
		if err != nil {
			return fmt.Errorf("unable to get signature value: failed decoding hash %s: %w", digest, err)
		}
	case MediaTypePEM:
		signaturePemBlocks, err := GetSignaturePEMBlocks([]byte(signature))
		if err != nil {
			return fmt.Errorf("unable to get signature pem blocks: %w", err)
		}
		if len(signaturePemBlocks) != 1 {
			return fmt.Errorf("expected 1 signature pem block, found %d", len(signaturePemBlocks))
		}
		signatureBytes = signaturePemBlocks[0].Bytes
	default:
		return fmt.Errorf("invalid signature mediaType %s", mediatype)
	}

	decodedHash, err := hex.DecodeString(digest)
	if err != nil {
		return fmt.Errorf("failed decoding hash %s: %w", digest, err)
	}
	// ensure length of hash is correct
	//if len(decodedHash) != 32 {
	//	return fmt.Errorf("hash to verify has invalid length")
	//}
	if err := rsa.VerifyPKCS1v15(publicKey, 0, decodedHash, signatureBytes); err != nil {
		return fmt.Errorf("signature verification failed, %w", err)
	}
	return nil
}

// GetSignaturePEMBlocks returns all signature pem blocks from a list of pem blocks
func GetSignaturePEMBlocks(pemData []byte) ([]*pem.Block, error) {
	if len(pemData) == 0 {
		return []*pem.Block{}, nil
	}

	signatureBlocks := []*pem.Block{}
	for {
		var currentBlock *pem.Block
		currentBlock, pemData = pem.Decode(pemData)
		if currentBlock == nil && len(pemData) > 0 {
			return nil, fmt.Errorf("unable to decode pem block %s", string(pemData))
		}

		if currentBlock.Type == SignaturePEMBlockType {
			signatureBlocks = append(signatureBlocks, currentBlock)
		}

		if len(pemData) == 0 {
			break
		}
	}

	return signatureBlocks, nil
}

func (_ Handler) CreateKeyPair() (priv interface{}, pub interface{}, err error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	return key, &key.PublicKey, nil
}
