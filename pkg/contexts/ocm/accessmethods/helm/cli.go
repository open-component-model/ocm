// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package helm

import (
	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/options"
)

func ConfigHandler() flagsets.ConfigOptionTypeSetHandler {
	return flagsets.NewConfigOptionTypeSetHandler(
		Type, AddConfig,
		options.RepositoryOption,
		options.PackageOption,
	)
}

func AddConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	flagsets.AddFieldByOptionP(opts, options.RepositoryOption, config, "helmRepository")
	flagsets.AddFieldByOptionP(opts, options.PackageOption, config, "helmChart")
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

- **<code>caCert</code>** *string*

  An optional TLS root certificate.

- **<code>keyring</code>** *string*

  An optional keyring used to verify the chart.
`
