package internal

import (
	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/modern-go/reflect2"
)

// CredentialsSource is a factory for effective credentials.
type CredentialsSource interface {
	Credentials(Context, ...CredentialsSource) (Credentials, error)
}

// CredentialsChain is a chain of credentials, where the
// credential i+1 (is present) is used to resolve credential i.
type CredentialsChain []CredentialsSource

var _ CredentialsSource = CredentialsChain{}

func (c CredentialsChain) Credentials(ctx Context, creds ...CredentialsSource) (Credentials, error) {
	if len(c) == 0 || reflect2.IsNil(c[0]) {
		return nil, nil
	}

	if len(creds) == 0 {
		return c[0].Credentials(ctx, c[1:]...)
	}
	return c[0].Credentials(ctx, sliceutils.CopyAppend(c[1:], creds...)...)
}
