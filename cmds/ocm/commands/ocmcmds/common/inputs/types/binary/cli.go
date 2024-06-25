package binary

import (
	"github.com/open-component-model/ocm/api/utils/cobrautils/flagsets"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/options"
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
