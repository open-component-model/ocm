package directory

import (
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/options"
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
