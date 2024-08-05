package signing

import (
	"encoding/hex"
	"hash"

	"github.com/mandelsoft/goutils/errors"
)

func Hash(hash hash.Hash, data []byte) (string, error) {
	hash.Reset()
	if _, err := hash.Write(data); err != nil {
		return "", errors.Wrapf(err, "failed hashing")
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
