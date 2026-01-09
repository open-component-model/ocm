package sigstore

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag/conv"
	"github.com/mandelsoft/goutils/errors"
	"github.com/sigstore/cosign/v3/cmd/cosign/cli/fulcio"
	"github.com/sigstore/cosign/v3/cmd/cosign/cli/options"
	"github.com/sigstore/cosign/v3/pkg/cosign"
	"github.com/sigstore/rekor/pkg/client"
	"github.com/sigstore/rekor/pkg/generated/client/entries"
	"github.com/sigstore/rekor/pkg/generated/models"
	hashedrekord_v001 "github.com/sigstore/rekor/pkg/types/hashedrekord/v0.0.1"
	"github.com/sigstore/rekor/pkg/verify"
	"github.com/sigstore/sigstore/pkg/signature"
	signatureoptions "github.com/sigstore/sigstore/pkg/signature/options"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/tech/signing"
	"ocm.software/ocm/api/tech/signing/handlers/sigstore/attr"
)

// Algorithm defines the type for the RSA PKCS #1 v1.5 signature algorithm.
// Since "sigstore" contained a buggy implementation, we are introducing "sigstore-v3" as well.
// "sigstore-v3" correctly injects the Fulcio fs.Cert into the bundle, not just the public key like in "sigstore".
const (
	Algorithm   = "sigstore"
	AlgorithmV3 = "sigstore-v3"
)

// MediaType defines the media type for a plain RSA signature.
const MediaType = "application/vnd.ocm.signature.sigstore"

func init() {
	signing.DefaultHandlerRegistry().RegisterSigner(Algorithm, Handler{algorithm: Algorithm})
	signing.DefaultHandlerRegistry().RegisterSigner(AlgorithmV3, Handler{algorithm: AlgorithmV3})
}

// Handler is a signatures.Signer compatible struct to sign using sigstore
// and a signatures.Verifier compatible struct to verify using sigstore.
// Uses "algorithm" field to distinguish between old "sigstore" and new "sigstore-v3" flows.
type Handler struct {
	algorithm string
}

// Algorithm specifies the name of the signing algorithm.
func (h Handler) Algorithm() string {
	return h.algorithm
}

// Sign implements the signing functionality.
// Since the "sigstore" algorithm had a buggy implementation, we are introducing "sigstore-v3" as well.
// We use the algorithm name to decide if old sigstore or new sigstore-v3 flow is used to
// guarantee backwards compatibility.
func (h Handler) Sign(cctx credentials.Context, digest string, sctx signing.SigningContext) (*signing.Signature, error) {
	hash := sctx.GetHash()
	// exit immediately if hash alg is not SHA-256, rekor doesn't currently support other hash functions
	if hash != crypto.SHA256 {
		return nil, fmt.Errorf("cannot sign using sigstore. rekor only supports SHA-256 digests: %s provided", hash.String())
	}

	ctx := context.Background()

	// generate a private key
	priv, err := cosign.GeneratePrivateKey()
	if err != nil {
		return nil, fmt.Errorf("error generating keypair: %w", err)
	}

	// create an ECDSA signer
	signer, err := signature.LoadECDSASignerVerifier(priv, hash)
	if err != nil {
		return nil, fmt.Errorf("error loading sigstore signer: %w", err)
	}

	// get the attributes for the sigstore signer
	cfg := attr.Get(cctx)

	// create a fulcio signing client
	fs, err := fulcio.NewSigner(ctx, options.KeyOpts{
		FulcioURL:        cfg.FulcioURL,
		OIDCIssuer:       cfg.OIDCIssuer,
		OIDCClientID:     cfg.OIDCClientID,
		SkipConfirmation: true,
	}, signer)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer: %w", err)
	}

	// decode the digest string
	rawDigest, err := hex.DecodeString(digest)
	if err != nil {
		return nil, fmt.Errorf("failed to decode digest: %w", err)
	}

	// sign the existing digest
	sig, err := fs.SignMessage(nil,
		signatureoptions.WithDigest(rawDigest),
		signatureoptions.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %w", err)
	}

	// get the public key for certificate transparency log
	pubKeys, err := cosign.GetCTLogPubs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get cosign CT Log Public Keys: %w", err)
	}

	// verify the signed certificate timestamp
	if err := cosign.VerifySCT(ctx, fs.Cert, fs.Chain, fs.SCT, pubKeys); err != nil {
		return nil, fmt.Errorf("failed to verify signed certificate timestamp: %w", err)
	}

	// get the public key from the signing key pair
	pub, err := fs.PublicKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get public key for signing: %w", err)
	}

	// marshal the public key bytes
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key for signing: %w", err)
	}

	// encode the public key to pem format
	publicKey := pem.EncodeToMemory(&pem.Block{
		Bytes: publicKeyBytes,
		Type:  "PUBLIC KEY",
	})

	// init the rekor client
	rekorClient, err := client.GetRekorClient(cfg.RekorURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create rekor client: %w", err)
	}

	// decide which public material to use for rekor entry
	// old "sigstore" flow uses only public key
	// new "sigstore-v3" flow uses the fulcio certificate
	// Since v3 the Fulcio certificate fs.Cert is already in PEM format
	rekorPublicMaterial := publicKey
	if h.Algorithm() == AlgorithmV3 {
		rekorPublicMaterial = fs.Cert
	}

	// create a rekor hashed entry
	hashedEntry := prepareRekorEntry(digest, sig, rekorPublicMaterial)

	// validate the rekor entry before submission
	if _, err := hashedEntry.Canonicalize(ctx); err != nil {
		return nil, fmt.Errorf("rekor entry is not valid: %w", err)
	}

	// prepare the entry for submission
	entry := &models.Hashedrekord{
		APIVersion: conv.Pointer(hashedEntry.APIVersion()),
		Spec:       hashedEntry.HashedRekordObj,
	}

	// prepare the create entry request parameters
	params := entries.NewCreateLogEntryParams().
		WithContext(ctx).
		WithProposedEntry(entry)

	// submit the create entry request
	resp, err := rekorClient.Entries.CreateLogEntry(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create rekor entry: %w", err)
	}

	// extract the payload from the rekor response
	data, err := json.Marshal(resp.GetPayload())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal rekor response: %w", err)
	}

	// store the rekor response in the signature value
	// depending on the used algorithm, either "sigstore" or "sigstore-v3"
	return &signing.Signature{
		Value:     base64.StdEncoding.EncodeToString(data),
		MediaType: MediaType,
		Algorithm: h.Algorithm(),
		Issuer:    "",
	}, nil
}

// Verify checks the signature, returns an error on verification failure.
func (h Handler) Verify(digest string, sig *signing.Signature, sctx signing.SigningContext) (err error) {
	ctx := context.Background()

	data, err := base64.StdEncoding.DecodeString(sig.Value)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	var entries models.LogEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("failed to unmarshal rekor log entry from signature: %w", err)
	}

	rawDigest, err := hex.DecodeString(digest)
	if err != nil {
		return fmt.Errorf("failed to decode digest: %w", err)
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

		rekorEntry := &models.Hashedrekord{}
		if err := json.Unmarshal(body, rekorEntry); err != nil {
			return fmt.Errorf("failed to unmarshal rekor entry body: %w", err)
		}

		if err := rekorEntry.Validate(strfmt.Default); err != nil {
			return fmt.Errorf("failed to validate rekor entry: %w", err)
		}

		rekorEntrySpec := rekorEntry.Spec.(map[string]any)
		rekorHashValue := rekorEntrySpec["data"].(map[string]any)["hash"].(map[string]any)["value"]

		// ensure digest matches
		if rekorHashValue != digest {
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

		block, _ := pem.Decode(rekorPublicKeyRaw)
		if block == nil {
			return fmt.Errorf("failed to decode public key: %w", err)
		}

		rekorPublicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse public key: %w", err)
		}

		// verify signature
		if err := ecdsa.VerifyASN1(rekorPublicKey.(*ecdsa.PublicKey), rawDigest, rekorSignature); !err {
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

// based on: https://github.com/sigstore/cosign/blob/ff648d5fb4ed6d0d1c16eaaceff970411fa969e3/pkg/cosign/tlog.go#L233
func prepareRekorEntry(digest string, sig, publicKey []byte) hashedrekord_v001.V001Entry {
	// TODO: this should match the provided hash digest algorithm but
	// rekor only supports SHA256 right now
	return hashedrekord_v001.V001Entry{
		HashedRekordObj: models.HashedrekordV001Schema{
			Data: &models.HashedrekordV001SchemaData{
				Hash: &models.HashedrekordV001SchemaDataHash{
					Algorithm: conv.Pointer(models.HashedrekordV001SchemaDataHashAlgorithmSha256),
					Value:     conv.Pointer(digest),
				},
			},
			Signature: &models.HashedrekordV001SchemaSignature{
				Content: strfmt.Base64(sig),
				PublicKey: &models.HashedrekordV001SchemaSignaturePublicKey{
					Content: strfmt.Base64(publicKey),
				},
			},
		},
	}
}

func extractECDSAPublicKey(pubKeyBytes []byte) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode(pubKeyBytes)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found in Fulcio public key")
	}
	switch block.Type {
	case "PUBLIC KEY":
		result, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Fulcio public key: %w", err)
		}
		// cast to ecdsa.PublicKey as we use this in the verification
		pub, ok := result.(*ecdsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("unexpected public key type: %T", result)
		}
		return pub, nil
	case "CERTIFICATE":
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Fulcio certificate: %w", err)
		}
		// cast to ecdsa.PublicKey as we use this in the verification
		pub, ok := cert.PublicKey.(*ecdsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("unexpected certificate public key type: %T", cert.PublicKey)
		}
		return pub, nil
	default:
		return nil, fmt.Errorf("unsupported PEM block type: %s", block.Type)
	}
}
