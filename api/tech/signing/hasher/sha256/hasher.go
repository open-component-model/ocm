package sha256

import (
	"crypto"
	"crypto/sha256"
	"hash"

	"ocm.software/ocm/api/tech/signing"
)

var Algorithm = crypto.SHA256.String()

func init() {
	signing.DefaultHandlerRegistry().RegisterHasher(Handler{})
}

// Handler is a signatures.Hasher compatible struct to hash with sha256.
type Handler struct{}

var _ signing.Hasher = Handler{}

func (_ Handler) Algorithm() string {
	return Algorithm
}

// Create creates a Hasher instance for no digest.
func (_ Handler) Create() hash.Hash {
	return sha256.New()
}

func (_ Handler) Crypto() crypto.Hash {
	return crypto.SHA256
}
