package directcreds

import (
	"ocm.software/ocm/api/credentials/cpi"
	"ocm.software/ocm/api/utils/misc"
)

func NewCredentials(props misc.Properties) cpi.CredentialsSpec {
	return cpi.NewCredentialsSpec(Type, NewRepositorySpec(props))
}
