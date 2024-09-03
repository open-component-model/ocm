package npm

import (
	"ocm.software/ocm/api/ocm/extensions/accessmethods/options"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
)

func ConfigHandler() flagsets.ConfigOptionTypeSetHandler {
	return flagsets.NewConfigOptionTypeSetHandler(
		Type, AddConfig,
		options.RepositoryOption,
		options.PackageOption,
		options.VersionOption,
	)
}

func AddConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	flagsets.AddFieldByOptionP(opts, options.RepositoryOption, config, "registry")
	flagsets.AddFieldByOptionP(opts, options.PackageOption, config, "package")
	flagsets.AddFieldByOptionP(opts, options.VersionOption, config, "version")
	return nil
}

var usage = `
This method implements the access of an NPM package in an NPM registry.
`

var formatV1 = `
The type specific specification fields are:

- **<code>registry</code>** *string*

  Base URL of the NPM registry.

- **<code>package</code>** *string*

  The name of the NPM package

- **<code>version</code>** *string*

  The version name of the NPM package
`
