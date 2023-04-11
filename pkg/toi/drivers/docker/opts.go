// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"github.com/docker/docker/api/types/container"
)

func NetworkModeOpt(mode string) ConfigurationOption {
	return func(_ *container.Config, h *container.HostConfig) error {
		h.NetworkMode = container.NetworkMode(mode)
		return nil
	}
}
