package internal

import (
	"crypto"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/generics"

	"ocm.software/ocm/api/credentials"
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/tech/signing"
	"ocm.software/ocm/api/tech/signing/signutils"
)

type (
	SignatureSpec = v1.SignatureSpec
)

func SignatureSpecFor(sig *signing.Signature) *SignatureSpec {
	return v1.SignatureSpecFor(sig)
}

type SigningContext struct {
	HashAlgo   crypto.Hash
	PrivateKey signutils.GenericPrivateKey
	PublicKey  signutils.GenericPublicKey
	Issuer     *pkix.Name

	Credentials credentials.DirectCredentials
}

var (
	_ json.Unmarshaler = (*SigningContext)(nil)
	_ json.Marshaler   = (*SigningContext)(nil)
)

type signingContext struct {
	HashAlgo    string                        `json:"hashAlgorithm"`
	PrivateKey  string                        `json:"privatekey,omitempty"`
	PublicKey   string                        `json:"publicKey,omitempty"`
	Issuer      string                        `json:"issuer,omitempty"`
	Credentials credentials.DirectCredentials `json:"credentials,omitempty"`
}

func (s SigningContext) MarshalJSON() ([]byte, error) {
	var err error

	ser := &signingContext{
		HashAlgo: s.HashAlgo.String(),
	}

	if s.PrivateKey != nil {
		ser.PrivateKey, err = encode(s.PrivateKey, signutils.GetPrivateKey, signutils.PemBlockForPrivateKey)
		if err != nil {
			return nil, err
		}
	}
	if s.PublicKey != nil {
		ser.PublicKey, err = encode(s.PublicKey, signutils.GetPublicKey, func(in interface{}) *pem.Block { return signutils.PemBlockForPublicKey(in) })
		if err != nil {
			return nil, err
		}
	}
	if s.Issuer != nil {
		ser.Issuer = signutils.DNAsString(*s.Issuer)
	}
	if s.Credentials != nil {
		ser.Credentials = s.Credentials
	}
	return json.Marshal(ser)
}

func (s *SigningContext) UnmarshalJSON(bytes []byte) error {
	var ser signingContext

	err := json.Unmarshal(bytes, &ser)
	if err != nil {
		return err
	}

	h := signing.DefaultRegistry().GetHasher(ser.HashAlgo)
	if h == nil {
		return errors.ErrUnknown(signutils.KIND_HASH_ALGORITHM, ser.HashAlgo)
	}
	s.HashAlgo = h.Crypto()
	if ser.Issuer != "" {
		s.Issuer, err = signutils.ParseDN(ser.Issuer)
		if err != nil {
			return err
		}
	}

	s.PrivateKey, err = decode[signutils.GenericPrivateKey](ser.PrivateKey, signutils.ParsePrivateKey)
	if err != nil {
		return err
	}
	s.PublicKey, err = decode[signutils.GenericPublicKey](ser.PublicKey, signutils.ParsePublicKey)
	if err != nil {
		return err
	}
	s.Credentials = ser.Credentials
	return nil
}

func encode[T any](in T, f func(in T) (interface{}, error), e func(in interface{}) *pem.Block) (string, error) {
	k, err := f(in)
	if err != nil {
		var i interface{} = in
		switch d := i.(type) {
		case []byte:
			return hex.EncodeToString(d), nil
		case string:
			return hex.EncodeToString([]byte(d)), nil
		default:
			return "", fmt.Errorf("invalid key type")
		}
	}
	b := e(k)
	if b == nil {
		return "", fmt.Errorf("cannot encode key")
	}
	return string(pem.EncodeToMemory(b)), nil
}

func decode[T any](in string, f func([]byte) (interface{}, error)) (T, error) {
	var _nil T
	k, err := f([]byte(in))
	if err != nil {
		data, err := hex.DecodeString(in)
		if err != nil {
			return _nil, err
		}
		k = data
	}
	return generics.Cast[T](k), nil
}
