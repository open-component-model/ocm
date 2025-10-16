package rsa_signingservice

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/tech/signing"
	"ocm.software/ocm/api/tech/signing/handlers/rsa"
	"ocm.software/ocm/api/utils/runtime"
)

// Algorithm defines the type for the RSA PKCS #1 v1.5 signature algorithm.
const (
	Algorithm = rsa.Algorithm
	Name      = "rsa-signingservice"
)

type Key struct {
	URL string `json:"url"`
}

// SignaturePEMBlockAlgorithmHeader defines the header in a signature pem block where the signature algorithm is defined.
const SignaturePEMBlockAlgorithmHeader = "Algorithm"

func init() {
	signing.DefaultHandlerRegistry().RegisterSigner(Name, NewHandler())
}

// Handler is a signatures.Signer compatible struct to sign with RSASSA-PKCS1-V1_5.
// using a signature service.
type Handler struct {
	algo string
}

func NewHandlerFor(algo string) signing.Signer {
	return &Handler{algo}
}

func NewHandler() signing.Signer {
	return &Handler{Algorithm}
}

func (h *Handler) Algorithm() string {
	return h.algo
}

func (h *Handler) Sign(cctx credentials.Context, digest string, sctx signing.SigningContext) (signature *signing.Signature, err error) {
	privateKey, err := PrivateKey(sctx.GetPrivateKey())
	if err != nil {
		return nil, errors.Wrapf(err, "invalid signing server access configuration")
	}
	server, err := NewSigningClient(privateKey.URL)
	if err != nil {
		return nil, err
	}
	return server.Sign(cctx, h.Algorithm(), sctx.GetHash(), digest, sctx)
}

func PrivateKey(k interface{}) (*Key, error) {
	switch t := k.(type) {
	case *Key:
		return t, nil
	case []byte:
		key := &Key{}
		err := runtime.DefaultYAMLEncoding.Unmarshal(t, key)
		if err != nil {
			return nil, err
		}
		return key, err
	default:
		return nil, fmt.Errorf("unknown key specification %T", k)
	}
}
