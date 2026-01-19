package sigstore

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"

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

// Helper to extract signature from descriptor YAML
func getSignatureByName(t *testing.T, descriptorYAML []byte, name string) (digest, sigValue string) {
	var descriptor map[string]any
	err := yaml.Unmarshal(descriptorYAML, &descriptor)
	require.NoError(t, err)

	sigs, ok := descriptor["signatures"].([]any)
	require.True(t, ok)

	for _, s := range sigs {
		sig := s.(map[string]any)
		if sig["name"].(string) == name {
			digest = sig["digest"].(map[string]any)["value"].(string)
			sigValue = sig["signature"].(map[string]any)["value"].(string)
			return
		}
	}
	t.Fatalf("signature %s not found", name)
	return
}

// ============================================================================
// Pure Function Tests (No Network, No OIDC flows)
// ============================================================================

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

	assert.NoError(t, err, "extractECDSAPublicKey should successfully parse valid PUBLIC KEY PEM block")
}

// Test extracting public key from certificate
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

	assert.NoError(t, err, "extractECDSAPublicKey should successfully extract public key from CERTIFICATE PEM block")
}

// Test error handling for invalid PEM
func TestExtractECDSAPublicKey_InvalidPEM(t *testing.T) {
	invalidPEM := []byte("not a valid pem block")

	_, err := extractECDSAPublicKey(invalidPEM)

	assert.Error(t, err)
}

// Test error handling for malformed certificate
func TestExtractECDSAPublicKey_MalformedCertificate(t *testing.T) {
	malformedCert := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: []byte("invalid certificate data"),
	})

	_, err := extractECDSAPublicKey(malformedCert)

	assert.Error(t, err)
}

// Test error handling for unsupported PEM type
func TestExtractECDSAPublicKey_UnsupportedPEMType(t *testing.T) {
	unsupportedPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "UNSUPPORTED",
		Bytes: []byte("some data"),
	})

	_, err := extractECDSAPublicKey(unsupportedPEM)

	assert.Error(t, err)
}

// ============================================================================
// Verify signatures with both Sigstore algorithms (backwards compatibility)
// These tests work OFFLINE because:
// - Rekor public keys are embedded in cosign library
// - All verification data is in the self-contained Sigstore bundle
// ============================================================================

// Test verifying legacy "sigstore" signature (created with old OCM CLI up to version v0.35.x)
// This ensures backwards compatibility - old signatures can be verified with new code
func TestVerify_LegacySignature(t *testing.T) {
	descriptorYAML := loadTestData(t, "component-descriptor-signed.yaml")
	digest, sigValue := getSignatureByName(t, descriptorYAML, "sigstore-legacy")

	handler := Handler{algorithm: Algorithm}

	err := handler.Verify(digest, &signing.Signature{
		Value:     sigValue,
		MediaType: MediaType,
		Algorithm: Algorithm,
	}, nil)

	assert.NoError(t, err, "legacy signature verification with new code should succeed")
}

// Test verifying "sigstore-v3" signature (created with new OCM CLI with fix)
// This tests the corrected implementation with Fulcio certificate in Rekor bundle
func TestVerify_V3Signature(t *testing.T) {

	descriptorYAML := loadTestData(t, "component-descriptor-signed.yaml")
	digest, sigValue := getSignatureByName(t, descriptorYAML, "sigstore-recommended")

	handler := Handler{algorithm: AlgorithmV3}

	err := handler.Verify(digest, &signing.Signature{
		Value:     sigValue,
		MediaType: MediaType,
		Algorithm: AlgorithmV3,
	}, nil)

	assert.NoError(t, err, "v3 signature verification should succeed")
}
