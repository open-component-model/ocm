package npm

import (
	dockercred "github.com/docker/cli/cli/config/credentials"

	"github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/oci/identity"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/utils"
)

const PROVIDER = "ocm.software/credentialprovider/" + REPOSITORY_TYPE

type ConsumerProvider struct {
	cfg string
}

var _ cpi.ConsumerProvider = (*ConsumerProvider)(nil)

func (p *ConsumerProvider) Unregister(id cpi.ProviderIdentity) {
}

func (p *ConsumerProvider) Match(req cpi.ConsumerIdentity, cur cpi.ConsumerIdentity, m cpi.IdentityMatcher) (cpi.CredentialsSource, cpi.ConsumerIdentity) {
	return p.get(req, cur, m)
}

func (p *ConsumerProvider) Get(req cpi.ConsumerIdentity) (cpi.CredentialsSource, bool) {
	creds, _ := p.get(req, nil, cpi.CompleteMatch)
	return creds, creds != nil
}

func (p *ConsumerProvider) get(req cpi.ConsumerIdentity, cur cpi.ConsumerIdentity, m cpi.IdentityMatcher) (cpi.CredentialsSource, cpi.ConsumerIdentity) {
	cfg := p.cfg
	all, err := ReadNpmConfigFile(cfg)
	if err != nil {
		panic(err)
		return nil, nil
	}

	var creds cpi.CredentialsSource

	for key, value := range all {
		hostname, port, _ := utils.SplitLocator(key)
		attrs := []string{identity.ID_HOSTNAME, hostname}
		if port != "" {
			attrs = append(attrs, identity.ID_PORT, port)
		}
		id := cpi.NewConsumerIdentity(identity.CONSUMER_TYPE, attrs...)
		if m(req, cur, id) {
			if IsEmptyAuthConfig(value) {
			} else {
				creds = newCredentials(value)
			}
			cur = id
		}
	}
	for h, helper := range cfg.CredentialHelpers {
		hostname := dockercred.ConvertToHostname(h)
		if hostname == "index.docker.io" {
			hostname = "docker.io"
		}
		id := cpi.ConsumerIdentity{
			cpi.ATTR_TYPE:        identity.CONSUMER_TYPE,
			identity.ID_HOSTNAME: hostname,
		}
		if m(req, cur, id) {
			creds = NewCredentials(cfg, h, dockercred.NewNativeStore(cfg, helper))
			cur = id
		}
	}
	return creds, cur
}
