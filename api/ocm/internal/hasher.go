package internal

import (
	"ocm.software/ocm/api/tech/signing"
)

// Hasher creates a new hash.Hash interface.
type Hasher = signing.Hasher

// HasherProvider provides access to supported hash methods.
type HasherProvider = signing.HasherProvider
