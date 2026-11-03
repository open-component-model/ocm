package git

import (
	"ocm.software/ocm/api/ocm/extensions/accessmethods/options"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
)

func ConfigHandler() flagsets.ConfigOptionTypeSetHandler {
	return configHandler(Type)
}

func UpperConfigHandler() flagsets.ConfigOptionTypeSetHandler {
	return configHandler(UpperType)
}

func configHandler(name string) flagsets.ConfigOptionTypeSetHandler {
	return flagsets.NewConfigOptionTypeSetHandler(
		name, AddConfig,
		options.RepositoryOption,
		options.ReferenceOption,
		options.CommitOption,
	)
}

func AddConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	flagsets.AddFieldByOptionP(opts, options.RepositoryOption, config, "repository")
	flagsets.AddFieldByOptionP(opts, options.CommitOption, config, "commit")
	flagsets.AddFieldByOptionP(opts, options.ReferenceOption, config, "ref")
	return nil
}

var usage = `
This method implements the access of the content of a git commit stored in a
Git repository. The blob is a gzip-compressed tar archive of the files checked
out for the commit.

This spelling and the version <code>v1alpha1</code> are legacy; they provide the
files only. Use the access type <code>Git</code>, whose archive also carries the
commit history and can be uploaded back into a Git repository by OCM v2.
`

var upperUsage = `
This method implements the access of the content of a git commit stored in a
Git repository. The blob is a gzip-compressed tar archive of the files checked
out for the commit and of a <code>.git</code> directory holding a detached
<code>HEAD</code> and the complete history reachable from that commit. It is
byte-identical to the archive OCM v2 builds for the same commit, so both
implementations digest and verify the same resource bytes.

The legacy spellings <code>git</code>, <code>git/v1alpha1</code> and
<code>Git/v1alpha1</code> provide the files only.
`

var formatV1 = `
The type specific specification fields are:

- **<code>repository</code>**  *string*

  Repository URL with or without scheme.

- **<code>ref</code>** (optional) *string*

  Original ref used to get the commit from

- **<code>commit</code>** *string*

  The sha/id of the git commit
`
