package ocm

import (
	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/ocm/cpi"
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
	flagsets.AddFieldByMappedOptionP(opts, options.RepositoryOption, config, repoMapper, "ocmRepository")
	flagsets.AddFieldByOptionP(opts, options.ComponentOption, config, "component")
	flagsets.AddFieldByOptionP(opts, options.VersionOption, config, "version")
	flagsets.AddFieldByOptionP(opts, options.IdentityPathOption, config, "resourceRef")
	return nil
}

func repoMapper(in any) (any, error) {
	uni, err := cpi.ParseRepo(in.(string))
	if err != nil {
		return nil, errors.ErrInvalidWrap(err, cpi.KIND_REPOSITORYSPEC, in.(string))
	}

	// TODO: basically a context is required, here.
	spec, err := cpi.DefaultContext().MapUniformRepositorySpec(&uni)
	if err != nil {
		return nil, err
	}
	return cpi.ToGenericRepositorySpec(spec)
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
