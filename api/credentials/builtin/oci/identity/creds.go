package identity

import (
	"path"

	"ocm.software/ocm/api/credentials/cpi"
	"ocm.software/ocm/api/utils"
)

func SimpleCredentials(user, passwd string) cpi.Credentials {
	return cpi.DirectCredentials{
		ATTR_USERNAME: user,
		ATTR_PASSWORD: passwd,
	}
}

func GetCredentials(ctx cpi.ContextProvider, locator, repo string) (cpi.Credentials, error) {
	return cpi.CredentialsForConsumer(ctx.CredentialsContext(), GetConsumerId(locator, repo), identityMatcher)
}

func GetConsumerId(locator, repo string) cpi.ConsumerIdentity {
	host, port, base := utils.SplitLocator(locator)
	id := cpi.NewConsumerIdentity(CONSUMER_TYPE, ID_HOSTNAME, host)
	if port != "" {
		id[ID_PORT] = port
	}
	if repo == "" {
		id[ID_PATHPREFIX] = base
	} else {
		id[ID_PATHPREFIX] = path.Join(base, repo)
	}
	return id
}
