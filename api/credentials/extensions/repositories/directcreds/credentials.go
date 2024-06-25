package directcreds

import (
	"github.com/open-component-model/ocm/api/common/common"
	"github.com/open-component-model/ocm/api/credentials/cpi"
)

func NewCredentials(props common.Properties) cpi.CredentialsSpec {
	return cpi.NewCredentialsSpec(Type, NewRepositorySpec(props))
}
