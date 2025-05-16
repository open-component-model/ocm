package github

import (
	"ocm.software/ocm/api/ocm/extensions/accessmethods/options"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
)

func ConfigHandler() flagsets.ConfigOptionTypeSetHandler {
	return flagsets.NewConfigOptionTypeSetHandler(
		Type, AddConfig,
		options.RepositoryOption,
		options.HostnameOption,
		options.CommitOption,
	)
}

func AddConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	flagsets.AddFieldByOptionP(opts, options.RepositoryOption, config, "repoUrl")
	flagsets.AddFieldByOptionP(opts, options.CommitOption, config, "commit")
	flagsets.AddFieldByOptionP(opts, options.HostnameOption, config, "apiHostname")
	return nil
}

var usage = `
This method implements the access of the content of a git commit stored in a
GitHub repository.
`

var formatV1 = `
The type specific specification fields are:

- **<code>repoUrl</code>**  *string*

  Repository URL with or without scheme.

- **<code>ref</code>** (optional) *string*

  Original ref used to get the commit from

- **<code>commit</code>** *string*

  The sha/id of the git commit
`
