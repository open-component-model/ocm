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
	flagsets.AddFieldByOptionP(opts, options.ReferenceOption, config, "localReference")
	flagsets.AddFieldByOptionP(opts, options.HintOption, config, "referenceName")
	flagsets.AddFieldByOptionP(opts, options.MediatypeOption, config, "mediaType")
	flagsets.AddFieldByOptionP(opts, options.GlobalAccessOption, config, "globalAccess")
	return nil
}
