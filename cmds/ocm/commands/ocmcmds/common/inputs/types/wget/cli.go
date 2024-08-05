package wget

import (
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/options"
)

func ConfigHandler() flagsets.ConfigOptionTypeSetHandler {
	return flagsets.NewConfigOptionTypeSetHandler(
		TYPE, AddConfig,
		options.URLOption,
		options.MediaTypeOption,
		options.HTTPHeaderOption,
		options.HTTPVerbOption,
		options.HTTPBodyOption,
		options.HTTPRedirectOption,
	)
}

func AddConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	flagsets.AddFieldByOptionP(opts, options.URLOption, config, "url")
	flagsets.AddFieldByOptionP(opts, options.MediaTypeOption, config, "mediaType")
	flagsets.AddFieldByOptionP(opts, options.HTTPHeaderOption, config, "header")
	flagsets.AddFieldByOptionP(opts, options.HTTPVerbOption, config, "verb")
	flagsets.AddFieldByOptionP(opts, options.HTTPBodyOption, config, "body")
	flagsets.AddFieldByOptionP(opts, options.HTTPRedirectOption, config, "noredirect")
	return nil
}
