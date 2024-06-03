package maven

import (
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/options"
	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
)

func ConfigHandler() flagsets.ConfigOptionTypeSetHandler {
	return flagsets.NewConfigOptionTypeSetHandler(
		TYPE, AddConfig,
		options.URLOption,
		options.PathOption,
		options.GroupOption,
		options.ArtifactOption,
		options.VersionOption,
		// optional
		options.ClassifierOption,
		options.ExtensionOption,
	)
}

func AddConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	flagsets.AddFieldByOptionP(opts, options.URLOption, config, "repoUrl")
	flagsets.AddFieldByOptionP(opts, options.PathOption, config, "path")
	flagsets.AddFieldByOptionP(opts, options.GroupOption, config, "groupId")
	flagsets.AddFieldByOptionP(opts, options.ArtifactOption, config, "artifactId")
	flagsets.AddFieldByOptionP(opts, options.VersionOption, config, "version")
	// optional
	flagsets.AddFieldByOptionP(opts, options.ClassifierOption, config, "classifier")
	flagsets.AddFieldByOptionP(opts, options.ExtensionOption, config, "extension")
	return nil
}
