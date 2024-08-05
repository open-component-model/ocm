package rsa_pss

import (
	"crypto"
	"crypto/rsa"
	"io"

	"ocm.software/ocm/api/tech/signing"
	rsahandler "ocm.software/ocm/api/tech/signing/handlers/rsa"
	"ocm.software/ocm/api/tech/signing/signutils"
)

// Algorithm defines the type for the RSA PKCS #1 v1.5 signature algorithm.
const Algorithm = "RSASSA-PSS"

// MediaType defines the media type for a plain RSA-PSS signature.
const MediaType = "application/vnd.ocm.signature.rsa.pss"

// MediaTypePEM is used if the signature contains the public key certificate chain.
const MediaTypePEM = signutils.MediaTypePEM

func init() {
	signing.DefaultHandlerRegistry().RegisterSigner(Algorithm, NewHandler())
}

func NewHandler() signing.SignatureHandler {
	return rsahandler.NewHandlerFor(RSASSA_PSS)
}

var RSASSA_PSS = &rsahandler.Method{
	Algorithm: Algorithm,
	MediaType: MediaType,
	Sign:      sign,
	Verify:    verify,
}

func sign(rand io.Reader, priv *rsa.PrivateKey, hash crypto.Hash, digest []byte) ([]byte, error) {
	return rsa.SignPSS(rand, priv, hash, digest, nil)
}

func verify(pub *rsa.PublicKey, hash crypto.Hash, digest []byte, sig []byte) error {
	return rsa.VerifyPSS(pub, hash, digest, sig, nil)
}
