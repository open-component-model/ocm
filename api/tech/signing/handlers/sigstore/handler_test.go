package sigstore

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/sigstore/rekor/pkg/generated/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"ocm.software/ocm/api/tech/signing"
)

// Helper function to load data from component descriptor
func loadTestData(t *testing.T, filename string) []byte {
	path := filepath.Join("testdata", filename)
	data, err := os.ReadFile(path)
	require.NoError(t, err, "failed to load test data: %s", filename)
	return data
}

// Helper to extract signature from descriptor YAML by algorithm
func getSignatureByAlgorithm(t *testing.T, descriptorYAML []byte, algorithm string) (digest, sigValue string) {
	var descriptor map[string]any
	err := yaml.Unmarshal(descriptorYAML, &descriptor)
	require.NoError(t, err)

	sigs, ok := descriptor["signatures"].([]any)
	require.True(t, ok)

	for _, s := range sigs {
		sig := s.(map[string]any)
		sigData := sig["signature"].(map[string]any)
		if sigData["algorithm"].(string) == algorithm {
			digest = sig["digest"].(map[string]any)["value"].(string)
			sigValue = sigData["value"].(string)
			return
		}
	}
	t.Fatalf("signature with algorithm %s not found", algorithm)
	return
}

// Test extracting public key from PEM format
func TestExtractECDSAPublicKey_FromPEMPublicKey(t *testing.T) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	require.NoError(t, err)

	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})
	require.NotNil(t, pemData)

	_, err = extractECDSAPublicKey(pemData)

	assert.NoError(t, err, "Should successfully parse valid PUBLIC KEY PEM block")
}

// Test extracting public key from cert
func TestExtractECDSAPublicKey_FromPEMCertificate(t *testing.T) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	certTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "test",
		},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, certTemplate, certTemplate, &privKey.PublicKey, privKey)
	require.NoError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	_, err = extractECDSAPublicKey(certPEM)

	assert.NoError(t, err, "Should successfully extract public key from CERTIFICATE PEM block")
}

// Test error handling for invalid PEM
func TestExtractECDSAPublicKey_InvalidPEM(t *testing.T) {
	invalidPEM := []byte("not a valid pem block")

	_, err := extractECDSAPublicKey(invalidPEM)

	assert.EqualError(t, err, "no PEM block found in Fulcio public key")
}

// Test error handling for malformed cert
func TestExtractECDSAPublicKey_MalformedCertificate(t *testing.T) {
	malformedCert := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: []byte("invalid certificate data"),
	})

	_, err := extractECDSAPublicKey(malformedCert)

	assert.ErrorContains(t, err, "failed to parse Fulcio certificate")
}

// Test error handling for unsupported PEM type
func TestExtractECDSAPublicKey_UnsupportedPEMType(t *testing.T) {
	unsupportedPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "UNSUPPORTED",
		Bytes: []byte("some data"),
	})

	_, err := extractECDSAPublicKey(unsupportedPEM)

	assert.EqualError(t, err, "unsupported PEM block type: UNSUPPORTED")
}

// Negative: Digest mismatch by passing wrong digest (no bundle mutation)
func TestVerify_DigestMismatch(t *testing.T) {
	descriptorYAML := loadTestData(t, "component-descriptor-signed.yaml")
	realDigest, sigValue := getSignatureByAlgorithm(t, descriptorYAML, AlgorithmV2)

	wrongDigest := "deadbeef" + realDigest

	handler := Handler{algorithm: AlgorithmV2}
	err := handler.Verify(wrongDigest, &signing.Signature{
		Value:     sigValue,
		MediaType: MediaType,
		Algorithm: AlgorithmV2,
	}, nil)

	assert.EqualError(t, err, "rekor hash doesn't match provided digest")
}

// Negative: Invalid signature bytes
func TestVerify_InvalidSignature(t *testing.T) {
	descriptorYAML := loadTestData(t, "component-descriptor-signed.yaml")
	digest, sigValue := getSignatureByAlgorithm(t, descriptorYAML, AlgorithmV2)

	// decode bundle
	var entries map[string]any
	data, err := base64.StdEncoding.DecodeString(sigValue)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, &entries))

	// mutate signature content of first entry
	for k, v := range entries {
		entry := v.(map[string]any)
		bodyB64 := entry["body"].(string)
		bodyJSONRaw, err := base64.StdEncoding.DecodeString(bodyB64)
		require.NoError(t, err)

		var rekorEntry models.Hashedrekord
		require.NoError(t, json.Unmarshal(bodyJSONRaw, &rekorEntry))

		rekorSpec := rekorEntry.Spec.(map[string]any)
		sigField := rekorSpec["signature"].(map[string]any)
		content := sigField["content"].(string)
		sigBytes, err := base64.StdEncoding.DecodeString(content)
		require.NoError(t, err)
		// flip one bit
		sigBytes[0] ^= 0x01
		sigField["content"] = base64.StdEncoding.EncodeToString(sigBytes)

		mutBody, err := json.Marshal(rekorEntry)
		require.NoError(t, err)
		entry["body"] = base64.StdEncoding.EncodeToString(mutBody)
		entries[k] = entry
		break
	}

	// re-encode bundle
	mutData, err := json.Marshal(entries)
	require.NoError(t, err)
	mutated := base64.StdEncoding.EncodeToString(mutData)

	// verify mutated signature
	handler := Handler{algorithm: AlgorithmV2}
	err = handler.Verify(digest, &signing.Signature{
		Value:     mutated,
		MediaType: MediaType,
		Algorithm: AlgorithmV2,
	}, nil)

	assert.EqualError(t, err, "could not verify signature using public key")
}

// Test handler for sigstore-v2 is registered and usable via signing registry
func TestHandlerRegistry_RegisteredAndUsable(t *testing.T) {
	verifier := signing.DefaultHandlerRegistry().GetVerifier(AlgorithmV2)
	require.NotNil(t, verifier, "v2 verifier should be registered")

	descriptorYAML := loadTestData(t, "component-descriptor-signed.yaml")
	digest, sigValue := getSignatureByAlgorithm(t, descriptorYAML, AlgorithmV2)

	err := verifier.Verify(digest, &signing.Signature{
		Value:     sigValue,
		MediaType: MediaType,
		Algorithm: AlgorithmV2,
	}, nil)
	assert.NoError(t, err, "verification via registry for v2 should succeed")
}

// Verify signatures with both Sigstore algorithms (works offline,
// as Rekor public keys are embedded in Cosign library and
// all verification data contained in Sigstore bundle)

// Verify legacy "sigstore" signature
func TestVerify_LegacySignature(t *testing.T) {
	descriptorYAML := loadTestData(t, "component-descriptor-signed.yaml")
	digest, sigValue := getSignatureByAlgorithm(t, descriptorYAML, Algorithm)

	handler := Handler{algorithm: Algorithm}

	err := handler.Verify(digest, &signing.Signature{
		Value:     sigValue,
		MediaType: MediaType,
		Algorithm: Algorithm,
	}, nil)

	assert.NoError(t, err, "legacy signature verification with new code should succeed")
}

// Verify "sigstore-v2" signature
func TestVerify_V2Signature(t *testing.T) {
	descriptorYAML := loadTestData(t, "component-descriptor-signed.yaml")
	digest, sigValue := getSignatureByAlgorithm(t, descriptorYAML, AlgorithmV2)

	handler := Handler{algorithm: AlgorithmV2}

	err := handler.Verify(digest, &signing.Signature{
		Value:     sigValue,
		MediaType: MediaType,
		Algorithm: AlgorithmV2,
	}, nil)

	assert.NoError(t, err, "v2 signature verification should succeed")
}
