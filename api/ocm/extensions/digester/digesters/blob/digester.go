package blob

import (
	"fmt"
	"io"

	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/tech/signing"
)

const GenericBlobDigestV1 = "genericBlobDigest/v1"

func init() {
	cpi.MustRegisterDigester(&defaultDigester{})
	cpi.SetDefaultDigester(&defaultDigester{})
}

type defaultDigester struct{}

var _ cpi.BlobDigester = (*defaultDigester)(nil)

func (d defaultDigester) GetType() cpi.DigesterType {
	return cpi.DigesterType{
		HashAlgorithm:          "",
		NormalizationAlgorithm: GenericBlobDigestV1,
	}
}

func (d defaultDigester) DetermineDigest(typ string, acc cpi.AccessMethod, preferred signing.Hasher) (*cpi.DigestDescriptor, error) {
	r, err := acc.Reader()
	if err != nil {
		return nil, err
	}
	hash := preferred.Create()

	if _, err := io.Copy(hash, r); err != nil {
		return nil, err
	}

	return &cpi.DigestDescriptor{
		Value:                  fmt.Sprintf("%x", hash.Sum(nil)),
		HashAlgorithm:          preferred.Algorithm(),
		NormalisationAlgorithm: GenericBlobDigestV1,
	}, nil
}
