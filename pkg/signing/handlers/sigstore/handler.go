// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package sigstore

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strings"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/sigstore/cosign/v2/cmd/cosign/cli/fulcio"
	"github.com/sigstore/cosign/v2/cmd/cosign/cli/options"
	"github.com/sigstore/cosign/v2/pkg/cosign"
	"github.com/sigstore/rekor/pkg/client"
	"github.com/sigstore/rekor/pkg/generated/client/entries"
	"github.com/sigstore/rekor/pkg/generated/models"
	"github.com/sigstore/rekor/pkg/types"
	"github.com/sigstore/rekor/pkg/types/rekord"
	"github.com/sigstore/rekor/pkg/verify"
	"github.com/sigstore/sigstore/pkg/signature"

	_ "github.com/sigstore/cosign/v2/pkg/providers/all"
	rekorv001 "github.com/sigstore/rekor/pkg/types/rekord/v0.0.1"
)

// Algorithm defines the type for the RSA PKCS #1 v1.5 signature algorithm.
const Algorithm = "sigstore"

// MediaType defines the media type for a plain RSA signature.
const MediaType = "application/vnd.ocm.signature.sigstore"

// SignaturePEMBlockAlgorithmHeader defines the header in a signature pem block where the signature algorithm is defined.
const SignaturePEMBlockAlgorithmHeader = "Algorithm"

func init() {
	signing.DefaultHandlerRegistry().RegisterSigner(Algorithm, Handler{})
}

// Handler is a signatures.Signer compatible struct to sign using sigstore
// and a signatures.Verifier compatible struct to verify using sigstore
type Handler struct{}

func (h Handler) Algorithm() string {
	return Algorithm
}

func (h Handler) Sign(cctx credentials.Context, digest string, hash crypto.Hash, issuer string, key interface{}) (*signing.Signature, error) {
	ctx := context.Background()

	priv, err := cosign.GeneratePrivateKey()
	if err != nil {
		return nil, fmt.Errorf("error generating keypair: %w", err)
	}

	signer, err := signature.LoadECDSASignerVerifier(priv, hash)
	if err != nil {
		return nil, fmt.Errorf("error loading sigstore signer: %w", err)
	}

	fs, err := fulcio.NewSigner(ctx, options.KeyOpts{
		FulcioURL:        "https://v1.fulcio.sigstore.dev",
		OIDCIssuer:       "https://oauth2.sigstore.dev/auth",
		OIDCClientID:     "sigstore",
		SkipConfirmation: true,
	}, signer)
	if err != nil {
		return nil, errors.Wrap(err, "new signer")
	}

	sig, err := fs.SignMessage(strings.NewReader(digest))
	if err != nil {
		return nil, fmt.Errorf("failed signing hash, %w", err)
	}

	pubKeys, err := cosign.GetCTLogPubs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get cosign CT Log Public Keys: %w", err)
	}

	if err := cosign.VerifySCT(ctx, fs.Cert, fs.Chain, fs.SCT, pubKeys); err != nil {
		return nil, fmt.Errorf("failed to verify signed certifcate timestamp: %w", err)
	}

	publicKey, err := fs.PublicKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get public key for signing: %w", err)
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key for signing: %w", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	rekorClient, err := client.GetRekorClient("https://rekor.sigstore.dev")
	if err != nil {
		return nil, fmt.Errorf("failed to create rekor client: %w", err)
	}

	rek := rekord.New()

	entry, err := rek.CreateProposedEntry(context.Background(), rekorv001.APIVERSION, types.ArtifactProperties{
		ArtifactBytes:  []byte(digest),
		SignatureBytes: sig,
		PublicKeyBytes: [][]byte{publicKeyPEM},
		PKIFormat:      "x509",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to prepare rekor entry: %w", err)
	}

	req := entries.NewCreateLogEntryParams().WithProposedEntry(entry)

	resp, err := rekorClient.Entries.CreateLogEntry(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create rekor entry: %w", err)
	}

	data, err := json.Marshal(resp.GetPayload())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal rekor response: %w", err)
	}

	return &signing.Signature{
		Value:     base64.StdEncoding.EncodeToString(data),
		MediaType: MediaType,
		Algorithm: Algorithm,
		Issuer:    issuer,
	}, nil
}

// Verify checks the signature, returns an error on verification failure.
func (h Handler) Verify(digest string, hash crypto.Hash, sig *signing.Signature, key interface{}) (err error) {
	ctx := context.Background()

	data, err := base64.StdEncoding.DecodeString(sig.Value)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	var entries models.LogEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("failed to unmarshal rekor log entry from signature: %w", err)
	}

	for _, entry := range entries {
		verifier, err := loadVerifier(ctx)
		if err != nil {
			return fmt.Errorf("failed to load rekor verifier: %w", err)
		}

		body, err := base64.StdEncoding.DecodeString(entry.Body.(string))
		if err != nil {
			return fmt.Errorf("failed to decode rekor entry body: %w", err)
		}

		rekorEntry := &models.Rekord{}
		if err := json.Unmarshal(body, rekorEntry); err != nil {
			return fmt.Errorf("failed to unmarshal rekor entry body: %w", err)
		}

		rekorEntrySpec := rekorEntry.Spec.(map[string]any)
		rekorHashValue := rekorEntrySpec["data"].(map[string]any)["hash"].(map[string]any)["value"]

		// ensure digest matches
		hashedDigest := hasher([]byte(digest))
		if rekorHashValue != hex.EncodeToString(hashedDigest) {
			return errors.New("rekor hash doesn't match provided digest")
		}

		// get the signature
		rekorSignatureRaw := rekorEntrySpec["signature"].(map[string]any)["content"]
		rekorSignature, err := base64.StdEncoding.DecodeString(rekorSignatureRaw.(string))
		if err != nil {
			return fmt.Errorf("failed to decode rekor signature: %w", err)
		}

		// get the public key from the signature
		rekorPublicKeyContent := rekorEntrySpec["signature"].(map[string]any)["publicKey"].(map[string]any)["content"]
		rekorPublicKeyRaw, err := base64.StdEncoding.DecodeString(rekorPublicKeyContent.(string))
		if err != nil {
			return fmt.Errorf("failed to decode rekor public key: %w", err)
		}

		block, _ := pem.Decode([]byte(rekorPublicKeyRaw))
		if block == nil {
			return fmt.Errorf("failed to decode public key: %w", err)
		}

		rekorPublicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse public key: %w", err)
		}

		// verify signature
		if err := ecdsa.VerifyASN1(rekorPublicKey.(*ecdsa.PublicKey), hashedDigest, rekorSignature); err != true {
			return errors.New("could not verify signature using public key")
		}

		// verify log entry
		if err := verify.VerifyLogEntry(ctx, &entry, verifier); err != nil {
			return fmt.Errorf("failed to verify log entry: %w", err)
		}
	}
	return nil
}

func loadVerifier(ctx context.Context) (signature.Verifier, error) {
	publicKeys, err := cosign.GetRekorPubs(ctx)
	if err != nil {
		return nil, err
	}

	for _, pubKey := range publicKeys.Keys {
		return signature.LoadVerifier(pubKey.PubKey, crypto.SHA256)
	}

	return nil, nil
}

func hasher(data []byte) []byte {
	hash := sha256.New()
	if _, err := hash.Write(data); err != nil {
		panic(err)
	}
	return hash.Sum(nil)
}
