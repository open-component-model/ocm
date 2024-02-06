// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package wget

import (
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/options"
	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
)

func ConfigHandler() flagsets.ConfigOptionTypeSetHandler {
	return flagsets.NewConfigOptionTypeSetHandler(
		TYPE, AddConfig,
		options.URLOption,
		options.MediaTypeOption,
		options.HTTPHeaderOption,
		options.HTTPVerbOption,
		options.HTTPBodyOption,
		options.HTTPRedirectOption,
	)
}

func AddConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	flagsets.AddFieldByOptionP(opts, options.URLOption, config, "url")
	flagsets.AddFieldByOptionP(opts, options.MediaTypeOption, config, "mediaType")
	flagsets.AddFieldByOptionP(opts, options.HTTPHeaderOption, config, "header")
	flagsets.AddFieldByOptionP(opts, options.HTTPVerbOption, config, "verb")
	flagsets.AddFieldByOptionP(opts, options.HTTPBodyOption, config, "body")
	flagsets.AddFieldByOptionP(opts, options.HTTPRedirectOption, config, "noredirect")
	return nil
}
