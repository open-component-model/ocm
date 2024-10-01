package plugin

import (
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/credentials/identity/hostpath"
	"ocm.software/ocm/api/ocm/plugin"
	"ocm.software/ocm/api/tech/signing"
)

type signer struct {
	plugin plugin.Plugin
	name   string
}

func NewSigner(p plugin.Plugin, name string) signing.Signer {
	return &signer{
		plugin: p,
		name:   name,
	}
}

func NewVerifier(p plugin.Plugin, name string) signing.Verifier {
	return &signer{
		plugin: p,
		name:   name,
	}
}

func (s *signer) Sign(ctx credentials.Context, digest string, sctx signing.SigningContext) (*signing.Signature, error) {
	cid, err := s.plugin.GetSigningConsumer(s.name, sctx)
	if err != nil {
		return nil, err
	}

	var creds credentials.DirectCredentials
	if len(cid) != 0 {
		c, err := credentials.CredentialsForConsumer(ctx, cid, hostpath.IdentityMatcher(cid.Type()))
		if err != nil {
			return nil, err
		}
		creds = credentials.DirectCredentials(c.Properties())
	}
	return s.plugin.Sign(s.name, digest, creds, sctx)
}

func (s *signer) Verify(digest string, sig *signing.Signature, sctx signing.SigningContext) error {
	return s.plugin.Verify(s.name, digest, sig, sctx)
}

func (s *signer) Algorithm() string {
	return s.name
}
