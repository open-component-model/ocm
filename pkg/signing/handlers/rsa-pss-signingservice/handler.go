// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package rsa_pss_signingservice

import (
	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa-pss"
	rsa_signingservice "github.com/open-component-model/ocm/pkg/signing/handlers/rsa-signingservice"
)

// Algorithm defines the type for the RSA PKCS #1 v1.5 signature algorithm.
const (
	Algorithm = rsa_pss.Algorithm
	Name      = "rsapss-signingservice"
)

// SignaturePEMBlockAlgorithmHeader defines the header in a signature pem block where the signature algorithm is defined.
const SignaturePEMBlockAlgorithmHeader = rsa_signingservice.SignaturePEMBlockAlgorithmHeader

func init() {
	signing.DefaultHandlerRegistry().RegisterSigner(Name, NewHandler())
}

func NewHandler() signing.Signer {
	return rsa_signingservice.NewHandlerFor(Algorithm)
}

type Key = rsa_signingservice.Key

func PrivateKey(k interface{}) (*Key, error) {
	return rsa_signingservice.PrivateKey(k)
}
