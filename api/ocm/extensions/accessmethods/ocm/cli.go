package ocm

import (
	"ocm.software/ocm/api/ocm/extensions/accessmethods/options"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
)

func ConfigHandler() flagsets.ConfigOptionTypeSetHandler {
	return flagsets.NewConfigOptionTypeSetHandler(
		Type, AddConfig,
		options.RepositoryOption,
		options.ComponentOption,
		options.VersionOption,
		options.IdentityPathOption,
	)
}

func AddConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	flagsets.AddFieldByMappedOptionP(opts, options.RepositoryOption, config, options.MapRepository, "ocmRepository")
	flagsets.AddFieldByOptionP(opts, options.ComponentOption, config, "component")
	flagsets.AddFieldByOptionP(opts, options.VersionOption, config, "version")
	flagsets.AddFieldByMappedOptionP(opts, options.IdentityPathOption, config, options.MapResourceRef, "resourceRef")
	return nil
}

var usage = `
This method implements the access of any resource artifact stored in an OCM
repository. Only repository types supporting remote access should be used.
`

var formatV1 = `
The type specific specification fields are:

- **<code>ocmRepository</code>** *json*

  The repository spec for the OCM repository

- **<code>component</code>** *string*

  *(Optional)* The name of the component. The default is the
  own component.

- **<code>version</code>** *string*

  *(Optional)* The version of the component. The default is the
  own component version.

- **<code>resourceRef</code>** *relative resource ref*

  The resource reference of the denoted resource relative to the
  given component version.

It uses the consumer identity and credentials for the intermediate repositories
and the final resource access.`
