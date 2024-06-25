package _default

import (
	"github.com/open-component-model/ocm/api/ocm/tools/toi/drivers/docker"
	"github.com/open-component-model/ocm/api/ocm/tools/toi/install"
)

var New = func() install.Driver {
	return &docker.Driver{}
}
