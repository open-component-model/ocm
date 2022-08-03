package repository

import (
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
)

type CredentialGetter struct {
	getCredentials func() (cpi.Credentials, error)
}

var _ cpi.CredentialsSource = CredentialGetter{}

func (c CredentialGetter) Credentials(ctx cpi.Context, cs ...cpi.CredentialsSource) (cpi.Credentials, error) {
	return c.getCredentials()
}
