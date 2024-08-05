package _default

import (
	"ocm.software/ocm/api/ocm/tools/toi/drivers/docker"
	"ocm.software/ocm/api/ocm/tools/toi/install"
)

var New = func() install.Driver {
	return &docker.Driver{}
}
