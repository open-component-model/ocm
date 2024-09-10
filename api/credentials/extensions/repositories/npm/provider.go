package npm

import (
	"ocm.software/ocm/api/credentials/cpi"
	npm "ocm.software/ocm/api/tech/npm/identity"
	"ocm.software/ocm/api/utils/logging"
)

type ConsumerProvider struct {
	npmrcPath string
}

var _ cpi.ConsumerProvider = (*ConsumerProvider)(nil)

func (p *ConsumerProvider) Unregister(_ cpi.ProviderIdentity) {
}

func (p *ConsumerProvider) Match(ectx cpi.EvaluationContext, req cpi.ConsumerIdentity, cur cpi.ConsumerIdentity, m cpi.IdentityMatcher) (cpi.CredentialsSource, cpi.ConsumerIdentity) {
	return p.get(req, cur, m)
}

func (p *ConsumerProvider) Get(req cpi.ConsumerIdentity) (cpi.CredentialsSource, bool) {
	creds, _ := p.get(req, nil, cpi.CompleteMatch)
	return creds, creds != nil
}

func (p *ConsumerProvider) get(requested cpi.ConsumerIdentity, currentFound cpi.ConsumerIdentity, m cpi.IdentityMatcher) (cpi.CredentialsSource, cpi.ConsumerIdentity) {
	all, path, err := readNpmConfigFile(p.npmrcPath)
	if err != nil {
		log := logging.Context().Logger(npm.REALM)
		log.LogError(err, "Failed to read npmrc file", "path", path)
		return nil, nil
	}

	var creds cpi.CredentialsSource

	for key, value := range all {
		id, err := npm.GetConsumerId("https://"+key, "")
		if err != nil {
			log := logging.Context().Logger(npm.REALM)
			log.LogError(err, "Failed to get consumer id", "key", key, "value", value)
			return nil, nil
		}
		if m(requested, currentFound, id) {
			creds = newCredentials(value)
			currentFound = id
		}
	}

	return creds, currentFound
}
