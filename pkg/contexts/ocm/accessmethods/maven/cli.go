package maven

import (
	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/options"
)

func ConfigHandler() flagsets.ConfigOptionTypeSetHandler {
	return flagsets.NewConfigOptionTypeSetHandler(
		Type, AddConfig,
		options.RepositoryOption,
		options.GroupOption,
		options.ArtifactOption,
		options.VersionOption,
		// optional
		options.ClassifierOption,
		options.ExtensionOption,
	)
}

func AddConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	flagsets.AddFieldByOptionP(opts, options.RepositoryOption, config, "repoUrl")
	flagsets.AddFieldByOptionP(opts, options.GroupOption, config, "groupId")
	flagsets.AddFieldByOptionP(opts, options.PackageOption, config, "artifactId")
	flagsets.AddFieldByOptionP(opts, options.VersionOption, config, "version")
	// optional
	flagsets.AddFieldByOptionP(opts, options.ClassifierOption, config, "classifier")
	flagsets.AddFieldByOptionP(opts, options.ExtensionOption, config, "extension")
	return nil
}

var usage = `
This method implements the access of a Maven artifact in a Maven repository.
`

var formatV1 = `
The type specific specification fields are:

- **<code>repoUrl</code>** *string*

  URL of the Maven repository

- **<code>groupId</code>** *string*

  The groupId of the Maven artifact

- **<code>artifactId</code>** *string*

  The artifactId of the Maven artifact

- **<code>version</code>** *string*

  The version name of the Maven artifact

- **<code>classifier</code>** *string*

  The optional classifier of the Maven artifact

- **<code>extension</code>** *string*

  The optional extension of the Maven artifact
`
