package ocm

import (
	acc "ocm.software/ocm/api/ocm/extensions/accessmethods/options"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/options"
)

func ConfigHandler() flagsets.ConfigOptionTypeSetHandler {
	return flagsets.NewConfigOptionTypeSetHandler(
		TYPE, AddConfig,
		options.RepositoryOption,
		options.ComponentOption,
		options.VersionOption,
		options.IdentityPathOption,
	)
}

func AddConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	flagsets.AddFieldByMappedOptionP(opts, options.RepositoryOption, config, acc.MapRepository, "ocmRepository")
	flagsets.AddFieldByOptionP(opts, options.ComponentOption, config, "component")
	flagsets.AddFieldByOptionP(opts, options.VersionOption, config, "version")
	flagsets.AddFieldByMappedOptionP(opts, options.IdentityPathOption, config, acc.MapResourceRef, "resourceRef")
	return nil
}
