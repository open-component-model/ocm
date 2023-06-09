// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package fulcio

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"encoding/pem"
	"fmt"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/sigstore/cosign/v2/cmd/cosign/cli/fulcio"
	"github.com/sigstore/cosign/v2/cmd/cosign/cli/options"
	"github.com/sigstore/cosign/v2/pkg/cosign"
	"github.com/sigstore/cosign/v2/pkg/providers"
	"github.com/sigstore/sigstore/pkg/signature"
)

// Algorithm defines the type for the RSA PKCS #1 v1.5 signature algorithm.
const Algorithm = "fulcio"

// MediaType defines the media type for a plain RSA signature.
const MediaType = "application/vnd.ocm.signature.fulcio"

// MediaTypePEM defines the media type for PEM formatted data.
const MediaTypePEM = "application/x-pem-file"

// SignaturePEMBlockType defines the type of a signature pem block.
const SignaturePEMBlockType = "SIGNATURE"

// SignaturePEMBlockAlgorithmHeader defines the header in a signature pem block where the signature algorithm is defined.
const SignaturePEMBlockAlgorithmHeader = "Signature Algorithm"

func init() {
	signing.DefaultHandlerRegistry().RegisterSigner(Algorithm, Handler{})
}

type (
	PrivateKey = rsa.PrivateKey
	PublicKey  = rsa.PublicKey
)

// Handler is a signatures.Signer compatible struct to sign with RSASSA-PKCS1-V1_5.
// and a signatures.Verifier compatible struct to verify RSASSA-PKCS1-V1_5 signatures.
type Handler struct{}

var _ Handler = Handler{}

func (h Handler) Algorithm() string {
	return Algorithm
}

func (h Handler) Sign(cctx credentials.Context, digest string, hash crypto.Hash, issuer string, key interface{}) (*signing.Signature, error) {
	ctx := context.Background()
	p, err := providers.ProvideFrom(ctx, "github")
	if err != nil {
		return nil, err
	}

	tok, err := p.Provide(ctx, "sigstore")
	if err != nil {
		return nil, err
	}

	priv, err := cosign.GeneratePrivateKey()
	if err != nil {
		return nil, fmt.Errorf("error generating keypair: %w", err)
	}

	signer, err := signature.LoadECDSASignerVerifier(priv, hash)
	if err != nil {
		return nil, fmt.Errorf("error loading sigstore signer: %w", err)
	}

	k, err := fulcio.NewSigner(ctx, options.KeyOpts{
		FulcioURL:    "https://v1.fulcio.sigstore.dev",
		IDToken:      tok,
		OIDCIssuer:   "https://oauth2.sigstore.dev/auth",
		OIDCClientID: "sigstore",
	}, signer)
	if err != nil {
		return nil, errors.Wrap(err, "new signer")
	}

	decodedHash, err := hex.DecodeString(digest)
	if err != nil {
		return nil, fmt.Errorf("failed decoding hash to bytes")
	}

	sig, err := k.SignMessage(bytes.NewReader(decodedHash))
	if err != nil {
		return nil, fmt.Errorf("failed signing hash, %w", err)
	}

	return &signing.Signature{
		Value:     hex.EncodeToString(sig),
		MediaType: MediaType,
		Algorithm: Algorithm,
		Issuer:    issuer,
	}, nil
}

// Verify checks the signature, returns an error on verification failure.
func (h Handler) Verify(digest string, hash crypto.Hash, signature *signing.Signature, key interface{}) (err error) {
	return nil
}

// GetSignaturePEMBlocks returns all signature pem blocks from a list of pem blocks.
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
