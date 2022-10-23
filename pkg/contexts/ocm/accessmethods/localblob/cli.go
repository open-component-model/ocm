//  SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
//  SPDX-License-Identifier: Apache-2.0

package localblob

import (
	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/options"
)

func ConfigHandler() flagsets.ConfigOptionTypeSetHandler {
	return flagsets.NewConfigOptionTypeSetHandler(
		Type, AddConfig,
		options.ReferenceOption,
		options.MediatypeOption,
		options.HintOption,
		options.GlobalAccessOption,
	)
}

func AddConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	if v, ok := opts.GetValue(options.ReferenceOption.Name()); ok {
		config["localReference"] = v
	}
	if v, ok := opts.GetValue(options.HintOption.Name()); ok {
		config["referenceName"] = v
	}
	if v, ok := opts.GetValue(options.MediatypeOption.Name()); ok {
		config["mediaType"] = v
	}
	if v, ok := opts.GetValue(options.GlobalAccessOption.Name()); ok {
		config["globalAccess"] = v
	}
	return nil
}
