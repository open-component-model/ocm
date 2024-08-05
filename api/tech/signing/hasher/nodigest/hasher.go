package nodigest

import (
	"crypto"
	"hash"

	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/tech/signing"
)

const Algorithm = metav1.NoDigest

func init() {
	signing.DefaultHandlerRegistry().RegisterHasher(Handler{})
}

// Handler is a signatures.Hasher compatible struct to hash with sha256.
type Handler struct{}

var _ signing.Hasher = Handler{}

func (h Handler) Algorithm() string {
	return Algorithm
}

// Create creates a Hasher instance for sha256.
func (_ Handler) Create() hash.Hash {
	return nil
}

func (_ Handler) Crypto() crypto.Hash {
	return 0
}
