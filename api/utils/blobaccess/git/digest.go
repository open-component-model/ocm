package git

import "github.com/opencontainers/go-digest"

type Digest interface {
	digest.Digest
}
