package npm

import (
	npm "github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/npm/identity"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/logging"
)

type ConsumerProvider struct {
	npmrcPath string
}

var _ cpi.ConsumerProvider = (*ConsumerProvider)(nil)

func (p *ConsumerProvider) Unregister(_ cpi.ProviderIdentity) {
}

func (p *ConsumerProvider) Match(req cpi.ConsumerIdentity, cur cpi.ConsumerIdentity, m cpi.IdentityMatcher) (cpi.CredentialsSource, cpi.ConsumerIdentity) {
	return p.get(req, cur, m)
}

func (p *ConsumerProvider) Get(req cpi.ConsumerIdentity) (cpi.CredentialsSource, bool) {
	creds, _ := p.get(req, nil, cpi.CompleteMatch)
	return creds, creds != nil
}

func (p *ConsumerProvider) get(requested cpi.ConsumerIdentity, currentFound cpi.ConsumerIdentity, m cpi.IdentityMatcher) (cpi.CredentialsSource, cpi.ConsumerIdentity) {
	all, err := readNpmConfigFile(p.npmrcPath)
	if err != nil {
		log := logging.Context().Logger(npm.REALM)
		log.LogError(err, "Failed to read npmrc file", "path", p.npmrcPath)
		return nil, nil
	}

	var creds cpi.CredentialsSource

	for key, value := range all {
		id := npm.GetConsumerId("https://"+key, "")

		if m(requested, currentFound, id) {
			creds = newCredentials(value)
			currentFound = id
		}
	}

	return creds, currentFound
}
