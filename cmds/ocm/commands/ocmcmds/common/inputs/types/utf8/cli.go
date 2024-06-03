package utf8

import (
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/options"
	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
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
