// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package attributes

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/listformat"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
)

func New(ctx clictx.Context) *cobra.Command {
	var standard credentials.IdentityMatcherInfos
	var consumer credentials.IdentityMatcherInfos
	for _, e := range ctx.CredentialsContext().ConsumerIdentityMatchers().List() {
		if e.IsConsumerType() {
			consumer = append(consumer, e)
		} else {
			standard = append(standard, e)
		}
	}

	return &cobra.Command{
		Use:   "credential-handling",
		Short: "Provisioning of credentials for credential consumers",
		Long: `
Because of the dynamic nature of the OCM area there are several kinds of
credential consumers with potentially completely different kinds of credentials.
Therefore, a common uniform credential management is required, capable to serve
all those use cases.

This is achieved by establishing a credential request mechanism based on
generic consumer identities and credential property sets.
On the one hand every kind of credential consumer uses a dedicated consumer
type (string). Additionally, it defines a set of properties further describing
the target/context credentials are required for.

On the other hand credentials can be defined for such sets of identities
with partial sets of properties (see <CMD>ocm configfile</CMD>). A credential
request is then matched against the available credential settings using matchers,
which might be specific for dedicated kinds of requests. For example, a hostpath
matcher matches a path prefix for a <code>pathprefix</code> property.

The best matching set of credential properties is then returned to the
credential consumer, which checks for the expected credential properties.

The following credential consumer types are used:
` + listformat.FormatListElements("", consumer) + `\
Those consumer types provide their own matchers, which are often based
on some standard generic matches. Those generic matchers and their
behaviours are described in the following list:
` + listformat.FormatListElements("", standard) + `
`,
	}
}
