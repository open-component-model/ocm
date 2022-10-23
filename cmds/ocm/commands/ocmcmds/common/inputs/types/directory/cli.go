//  SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
//  SPDX-License-Identifier: Apache-2.0

package directory

import (
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/options"
	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
)

func ConfigHandler() flagsets.ConfigOptionTypeSetHandler {
	return cpi.NewMediaFileSpecOptionType(
		TYPE, AddConfig,
		options.IncludeOption,
		options.ExcludeOption,
		options.PreserveDirOption,
		options.FollowSymlinksOption,
	)
}

func AddConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	if err := cpi.AddMediaFileSpecConfig(opts, config); err != nil {
		return err
	}
	if v, ok := opts.GetValue(options.PreserveDirOption.Name()); ok {
		config["preserveDir"] = v
	}
	if v, ok := opts.GetValue(options.FollowSymlinksOption.Name()); ok {
		config["followSymlinks"] = v
	}
	if v, ok := opts.GetValue(options.ExcludeOption.Name()); ok {
		config["excludeFiles"] = v
	}
	if v, ok := opts.GetValue(options.IncludeOption.Name()); ok {
		config["includeFiles"] = v
	}
	return nil
}
