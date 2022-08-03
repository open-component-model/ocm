package container_registry

import (
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	gardenercfg_cpi "github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/gardenerconfig/cpi"
)

type credentials struct {
	name             string
	consumerIdentity cpi.ConsumerIdentity
	data             cpi.Credentials
}

func (c credentials) Name() string {
	return c.name
}

func (c credentials) ConsumerIdentity() cpi.ConsumerIdentity {
	return c.consumerIdentity
}
func (c credentials) Data() cpi.Credentials {
	return c.data
}

var _ gardenercfg_cpi.Credential = credentials{}
