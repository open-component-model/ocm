package container_registry

import (
	"ocm.software/ocm/api/credentials/cpi"
	gardenercfgcpi "ocm.software/ocm/api/credentials/extensions/repositories/gardenerconfig/cpi"
)

type credentials struct {
	name             string
	consumerIdentity cpi.ConsumerIdentity
	properties       cpi.Credentials
}

func (c credentials) Name() string {
	return c.name
}

func (c credentials) ConsumerIdentity() cpi.ConsumerIdentity {
	return c.consumerIdentity
}

func (c credentials) Properties() cpi.Credentials {
	return c.properties
}

var _ gardenercfgcpi.Credential = credentials{}
