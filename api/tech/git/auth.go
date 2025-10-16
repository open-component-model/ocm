package git

import (
	"errors"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	gssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/tech/git/identity"
)

var ErrNoValidGitCredentials = errors.New("no valid credentials found for git authentication")

type AuthMethod = transport.AuthMethod

// AuthFromCredentials creates a git authentication method from the given credentials.
// If no valid credentials are found, ErrNoValidGitCredentials is returned.
// However, one can still perform anonymous operations with the git client if the repo allows it.
func AuthFromCredentials(creds credentials.Credentials) (AuthMethod, error) {
	if creds == nil {
		return nil, ErrNoValidGitCredentials
	}

	if creds.ExistsProperty(identity.ATTR_PRIVATE_KEY) {
		return gssh.NewPublicKeysFromFile(
			creds.GetProperty(identity.ATTR_USERNAME),
			creds.GetProperty(identity.ATTR_PRIVATE_KEY),
			creds.GetProperty(identity.ATTR_PASSWORD),
		)
	}

	if creds.ExistsProperty(identity.ATTR_TOKEN) {
		return &http.TokenAuth{Token: creds.GetProperty(identity.ATTR_TOKEN)}, nil
	}

	if creds.ExistsProperty(identity.ATTR_USERNAME) {
		return &http.BasicAuth{
			Username: creds.GetProperty(identity.ATTR_USERNAME),
			Password: creds.GetProperty(identity.ATTR_PASSWORD),
		}, nil
	}

	return nil, ErrNoValidGitCredentials
}
