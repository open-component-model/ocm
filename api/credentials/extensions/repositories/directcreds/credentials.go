package directcreds

import (
	"ocm.software/ocm/api/credentials/cpi"
	common "ocm.software/ocm/api/utils/misc"
)

func NewCredentials(props common.Properties) cpi.CredentialsSpec {
	return cpi.NewCredentialsSpec(Type, NewRepositorySpec(props))
}
