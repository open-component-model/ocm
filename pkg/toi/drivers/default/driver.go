package _default

import (
	"github.com/open-component-model/ocm/pkg/toi/drivers/docker"
	"github.com/open-component-model/ocm/pkg/toi/install"
)

var New = func() install.Driver {
	return &docker.Driver{}
}
