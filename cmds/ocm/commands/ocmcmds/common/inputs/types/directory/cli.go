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

	flagsets.AddFieldByOptionP(opts, options.PreserveDirOption, config, "preserveDir")
	flagsets.AddFieldByOptionP(opts, options.FollowSymlinksOption, config, "followSymlinks")
	flagsets.AddFieldByOptionP(opts, options.ExcludeOption, config, "excludeFiles")
	flagsets.AddFieldByOptionP(opts, options.IncludeOption, config, "includeFiles")
	return nil
}
