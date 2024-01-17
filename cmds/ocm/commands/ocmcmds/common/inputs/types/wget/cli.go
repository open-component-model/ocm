// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package directory

import (
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/options"
	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
)

func ConfigHandler() flagsets.ConfigOptionTypeSetHandler {
	return flagsets.NewConfigOptionTypeSetHandler(
		TYPE, AddConfig,
		options.URLOption,
		options.MediaTypeOption,
	)
}

func AddConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	flagsets.AddFieldByOptionP(opts, options.URLOption, config, "url")
	flagsets.AddFieldByOptionP(opts, options.MediaTypeOption, config, "mime type")
	return nil
}
