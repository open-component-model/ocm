package mvn

import (
	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/options"
)

func ConfigHandler() flagsets.ConfigOptionTypeSetHandler {
	return flagsets.NewConfigOptionTypeSetHandler(
		Type, AddConfig,
		options.RegistryOption,
		options.PackageOption,
		options.VersionOption,
	)
}

func AddConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	flagsets.AddFieldByOptionP(opts, options.RepositoryOption, config, "repository")
	flagsets.AddFieldByOptionP(opts, options.PackageOption, config, "package")
	flagsets.AddFieldByOptionP(opts, options.VersionOption, config, "version")
	return nil
}

var usage = `
This method implements the access of a Maven (mvn) package in a Maven repository.
`

var formatV1 = `
The type specific specification fields are:

- **<code>repository</code>** *string*

  Base URL of the Maven (mvn) repository.

- **<code>package</code>** *string*

  The name of the Maven (mvn) package

- **<code>version</code>** *string*

  The version name of the Maven (mvn) package
`
