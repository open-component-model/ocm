package ociartifact

import (
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/options"
)

func ConfigHandler(t string) flagsets.ConfigOptionTypeSetHandler {
	return cpi.NewMediaFileSpecOptionType(t, AddConfig,
		options.PathOption, options.HintOption, options.PlatformsOption)
}

func AddConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	if err := cpi.AddPathSpecConfig(opts, config); err != nil {
		return err
	}
	flagsets.AddFieldByOptionP(opts, options.HintOption, config, "repository")
	flagsets.AddFieldByOptionP(opts, options.PlatformsOption, config, "platforms")
	return nil
}
