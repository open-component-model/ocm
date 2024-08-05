package binary

import (
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/options"
)

func ConfigHandler() flagsets.ConfigOptionTypeSetHandler {
	set := flagsets.NewConfigOptionTypeSetHandler(TYPE, AddConfig, options.DataOption)
	cpi.AddProcessSpecOptionTypes(set)
	return set
}

func AddConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	flagsets.AddFieldByOptionP(opts, options.DataOption, config, "data")
	return cpi.AddProcessSpecConfig(opts, config)
}
