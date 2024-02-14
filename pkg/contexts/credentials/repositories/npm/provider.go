package npm

import (
	"net/url"

	"github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/oci/identity"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
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
		panic(err)
		return nil, nil
	}

	var creds cpi.CredentialsSource

	for key, value := range all {
		u, e := url.Parse(key)
		if e != nil {
			return nil, nil
		}

		attrs := []string{identity.ID_HOSTNAME, u.Hostname()}
		if u.Port() != "" {
			attrs = append(attrs, identity.ID_PORT, u.Port())
		}
		id := cpi.NewConsumerIdentity(identity.CONSUMER_TYPE, attrs...)
		if m(requested, currentFound, id) {
			creds = newCredentials(value)
			currentFound = id
		}
	}

	return creds, currentFound
}
