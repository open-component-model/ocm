//  SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
//  SPDX-License-Identifier: Apache-2.0

package ociartefact

import (
	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/options"
)

func ConfigHandler() flagsets.ConfigOptionTypeSetHandler {
	return flagsets.NewConfigOptionTypeSetHandler(
		Type, AddConfig,
		options.ReferenceOption,
	)
}

func AddConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	if v, ok := opts.GetValue(options.ReferenceOption.Name()); ok {
		config["imageReference"] = v
	}
	return nil
}
