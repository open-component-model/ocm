package signinghandlers

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/tech/signing"
	"ocm.software/ocm/api/tech/signing/signutils"
	"ocm.software/ocm/api/utils/mime"
)

const (
	NAME = "demo"
)

const (
	CID_TYPE   = "TEST"
	SUFFIX     = "suffix"
	SUFFIX_KEY = "suffix"
)

func New() ppi.SigningHandler {
	return ppi.NewSigningHandler(NAME, "fake signing", &signer{}).WithVerifier(&verifier{}).WithCredentials(provider)
}

////////////////////////////////////////////////////////////////////////////////

func provider(sctx signing.SigningContext) credentials.ConsumerIdentity {
	key, ok := sctx.GetPrivateKey().([]byte)
	if !ok {
		return nil
	}
	if string(key) == SUFFIX {
		return credentials.NewConsumerIdentity(CID_TYPE, "host", "localhost")
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type signer struct{}

var _ signing.Signer = (*signer)(nil)

func (s signer) Sign(cctx credentials.Context, digest string, sctx signing.SigningContext) (*signing.Signature, error) {
	var suffix string

	key, ok := sctx.GetPrivateKey().([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid private key")
	}

	if cctx != nil {
		c, err := credentials.CredentialsForConsumer(cctx, provider(sctx))
		if err != nil {
			return nil, err
		}
		if c != nil {
			suffix = c.GetProperty(SUFFIX_KEY)
			if suffix != "" {
				suffix = ":" + suffix
			}
		}
	}

	i := ""
	if sctx.GetIssuer() != nil {
		i = signutils.DNAsString(*sctx.GetIssuer())
	}
	if sctx.GetPrivateKey() == nil {
		return nil, fmt.Errorf("private key required")
	}
	return &signing.Signature{
		Value:     digest + ":" + string(key) + suffix,
		MediaType: mime.MIME_TEXT,
		Algorithm: s.Algorithm(),
		Issuer:    i,
	}, nil
}

func (s signer) Algorithm() string {
	return NAME
}

////////////////////////////////////////////////////////////////////////////////

type verifier struct{}

var _ signing.Verifier = (*verifier)(nil)

func (v verifier) Verify(digest string, sig *signing.Signature, sctx signing.SigningContext) error {
	if sig.Algorithm != NAME {
		return errors.ErrInvalid(signutils.KIND_SIGN_ALGORITHM, sig.Algorithm)
	}
	if sctx.GetPublicKey() == nil {
		return fmt.Errorf("public key required")
	}
	key, ok := sctx.GetPublicKey().([]byte)
	if !ok {
		return fmt.Errorf("invalid private key")
	}
	if sig.Value != digest+":"+string(key) {
		return fmt.Errorf("invalid signature %q != %q", sig.Value, digest+":"+string(key))
	}
	return nil
}

func (v verifier) Algorithm() string {
	return NAME
}
