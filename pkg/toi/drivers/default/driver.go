// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package _default

import (
	"github.com/open-component-model/ocm/v2/pkg/toi/drivers/docker"
	"github.com/open-component-model/ocm/v2/pkg/toi/install"
)

var New = func() install.Driver {
	return &docker.Driver{}
}
