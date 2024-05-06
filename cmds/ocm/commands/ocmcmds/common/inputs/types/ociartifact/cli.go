package ociartifact

import (
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/options"
	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
)

func ConfigHandler() flagsets.ConfigOptionTypeSetHandler {
	return cpi.NewMediaFileSpecOptionType(TYPE, AddConfig,
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
