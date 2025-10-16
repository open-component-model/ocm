package blobaccess

import (
	"github.com/opencontainers/go-digest"

	"ocm.software/ocm/api/utils/blobaccess/bpi"
)

func Digest(access bpi.DataAccess) (digest.Digest, error) {
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
