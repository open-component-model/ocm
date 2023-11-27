// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package signutils

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// MediaTypePEM defines the media type for PEM formatted data.
const MediaTypePEM = "application/x-pem-file"

// SignaturePEMBlockType defines the type of a signature pem block.
const SignaturePEMBlockType = "SIGNATURE"

// SignaturePEMBlockAlgorithmHeader defines the header in a signature pem block where the signature algorithm is defined.
const SignaturePEMBlockAlgorithmHeader = "Signature Algorithm"

// GetSignatureFromPem returns a signature and certificated contained
// in a PEM block list.
func GetSignatureFromPem(pemData []byte) ([]byte, string, []*x509.Certificate, error) {
	var signature []byte
	var algo string

	if len(pemData) == 0 {
		return nil, "", nil, nil
	}

	var currentBlock *pem.Block
	currentBlock, pemData = pem.Decode(pemData)
	if currentBlock == nil && len(pemData) > 0 {
		return nil, "", nil, fmt.Errorf("unable to decode pem block %s", string(pemData))
	}

	if currentBlock.Type == SignaturePEMBlockType {
		signature = currentBlock.Bytes
		algo = currentBlock.Headers[SignaturePEMBlockAlgorithmHeader]
	}

	caChain, err := ParseCertificateChain(pemData, false)
	if err != nil {
		return nil, "", nil, err
	}
	return signature, algo, caChain, nil
}

func SignatureBytesToPem(algo string, data []byte, certs ...*x509.Certificate) []byte {
	block := &pem.Block{Type: "CERTIFICATE", Bytes: data}
	if algo != "" {
		block.Headers = map[string]string{
			SignaturePEMBlockAlgorithmHeader: algo,
		}
	}
	return append(pem.EncodeToMemory(block), CertificateChainToPem(certs)...)
}
