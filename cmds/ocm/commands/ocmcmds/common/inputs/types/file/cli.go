package file

import (
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
)

func ConfigHandler() flagsets.ConfigOptionTypeSetHandler {
	return cpi.NewMediaFileSpecOptionType(TYPE, AddConfig)
}

func AddConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	return cpi.AddMediaFileSpecConfig(opts, config)
}
