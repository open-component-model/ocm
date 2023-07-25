// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/plugin/ppi"
)

func TweakDescriptor(d ppi.Descriptor, cfg *Config) ppi.Descriptor {
	if cfg != nil {
		d.Actions[0].DefaultSelectors = append(d.Actions[0].DefaultSelectors, cfg.Hostnames...)
	}
	return d
}
