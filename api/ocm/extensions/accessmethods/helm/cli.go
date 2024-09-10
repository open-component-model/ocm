package helm

import (
	"ocm.software/ocm/api/ocm/extensions/accessmethods/options"
	"ocm.software/ocm/api/tech/helm/identity"
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
	flagsets.AddFieldByOptionP(opts, options.RepositoryOption, config, "helmRepository")
	flagsets.AddFieldByOptionP(opts, options.PackageOption, config, "helmChart")
	flagsets.AddFieldByOptionP(opts, options.VersionOption, config, "version")
	return nil
}

var usage = `
This method implements the access of a Helm chart stored in a Helm repository.
`

var formatV1 = `
The type specific specification fields are:

- **<code>helmRepository</code>** *string*

  Helm repository URL.

- **<code>helmChart</code>** *string*

  The name of the Helm chart and its version separated by a colon.

- **<code>version</code>** *string*

  The version of the Helm chart if not specified as part of the chart name.

- **<code>caCert</code>** *string*

  An optional TLS root certificate.

- **<code>keyring</code>** *string*

  An optional keyring used to verify the chart.

It uses the consumer identity type ` + identity.CONSUMER_TYPE + ` with the fields
for a hostpath identity matcher (see <CMD>ocm get credentials</CMD>).`
