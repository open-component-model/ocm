package internal

import (
	"github.com/mandelsoft/goutils/errors"
)

func CredentialsForConsumer(ctx ContextProvider, id ConsumerIdentity, unknownAsError bool, matchers ...IdentityMatcher) (Credentials, error) {
	cctx := ctx.CredentialsContext()

	src, err := cctx.GetCredentialsForConsumer(id, matchers...)
	if err != nil {
		if !errors.IsErrUnknown(err) {
			return nil, errors.Wrapf(err, "lookup credentials failed for %s", id)
		}
		if unknownAsError {
			return nil, err
		}
		return nil, nil
	}
	creds, err := src.Credentials(cctx)
	if err != nil {
		unknownErr := errors.ErrUnknown(KIND_CREDENTIALS, id.String())
		err = errors.Wrapf(err, "unable to receive credentials for %s", id)
		return nil, errors.Wrap(unknownErr, err)
	}
	return creds, nil
}
