package github

import (
	"os"

	"ocm.software/ocm/api/credentials/cpi"
	"ocm.software/ocm/api/tech/github/identity"
	"ocm.software/ocm/api/utils/misc"
)

func init() {
	t := os.Getenv("GITHUB_TOKEN")
	if t != "" {
		us := os.Getenv("GITHUB_SERVER_URL")
		id := identity.GetConsumerId(us)

		if src, err := cpi.DefaultContext.GetCredentialsForConsumer(id); err != nil || src == nil {
			creds := cpi.NewCredentials(misc.Properties{cpi.ATTR_TOKEN: t})
			cpi.DefaultContext.SetCredentialsForConsumer(id, creds)
		}
	}
}
