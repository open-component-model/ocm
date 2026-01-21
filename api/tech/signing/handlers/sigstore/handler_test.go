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
