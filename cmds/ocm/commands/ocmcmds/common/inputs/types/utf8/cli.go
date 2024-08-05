package utf8

import (
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/options"
)

func ConfigHandler() flagsets.ConfigOptionTypeSetHandler {
	set := flagsets.NewConfigOptionTypeSetHandler(TYPE, AddConfig,
		options.TextOption, options.JSONOption, options.FormattedJSONOption, options.YAMLOption)
	cpi.AddProcessSpecOptionTypes(set)
	return set
}

func AddConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	flagsets.AddFieldByOptionP(opts, options.TextOption, config, "text")
	flagsets.AddFieldByOptionP(opts, options.JSONOption, config, "json")
	flagsets.AddFieldByOptionP(opts, options.FormattedJSONOption, config, "formattedJson")
	flagsets.AddFieldByOptionP(opts, options.YAMLOption, config, "yaml")
	return cpi.AddProcessSpecConfig(opts, config)
}
