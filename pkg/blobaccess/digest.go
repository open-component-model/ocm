package blobaccess

import (
	"github.com/opencontainers/go-digest"
)

func Digest(access DataAccess) (digest.Digest, error) {
	reader, err := access.Reader()
	if err != nil {
		return "", err
	}
	defer reader.Close()

	dig, err := digest.FromReader(reader)
	if err != nil {
		return "", err
	}
	return dig, nil
}
