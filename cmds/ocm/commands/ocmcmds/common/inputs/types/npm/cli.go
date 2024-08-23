package npm

import (
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/options"
)

func ConfigHandler() flagsets.ConfigOptionTypeSetHandler {
	return flagsets.NewConfigOptionTypeSetHandler(
		TYPE, AddConfig,
		options.RegistryOption,
		options.PackageOption,
		options.PackageVersionOption,
	)
}

func AddConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	flagsets.AddFieldByOptionP(opts, options.RegistryOption, config, "registry")
	flagsets.AddFieldByOptionP(opts, options.PackageOption, config, "package")
	flagsets.AddFieldByOptionP(opts, options.PackageVersionOption, config, "version")
	return nil
}
