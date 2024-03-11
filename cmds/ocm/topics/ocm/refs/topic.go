// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package topicocmrefs

import (
	"github.com/spf13/cobra"

	topicocirefs "github.com/open-component-model/ocm/cmds/ocm/topics/oci/refs"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

func New(ctx clictx.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "ocm-references",
		Short: "notation for OCM references",
		Example: `
+ctf+directory::./ocm/ctf//ocm.software/ocmcli:0.7.0

oci::{"baseUrl":"ghcr.io","componentNameMapping":"urlPath","subPath":"open-component-model"}//ocm.software/ocmcli.0.7.0

oci::https://ghcr.io:443/open-component-model//ocm.software/ocmcli:0.7.0

oci::http://localhost:8080/local-component-repository//ocm.software/ocmcli:0.7.0
`,
		Long: `
The command line client supports a special notation scheme for specifying
references to OCM components and repositories. This allows for specifying
references to any registry supported by the OCM toolset that can host OCM
components:

<center>
    <pre>[+][&lt;type>::][./]&lt;file path>//&lt;component id>[:&lt;version>]</pre>
        or
	<pre>[+][&lt;type>::][&lt;json repo spec>//]&lt;component id>[:&lt;version>]</pre>
		or
    <pre>[+][&lt;type>::][&lt;scheme>://]&lt;domain>[:&lt;port>][/&lt;repository prefix>]//&lt;component id>[:&lt;version]</pre>
		or
	<pre>[+][&lt;type>::][&lt;scheme>://]&lt;host>[:&lt;port>][/&lt;repository prefix>]//&lt;component id>[:&lt;version]</pre>
</center>

Besides dedicated components it is also possible to denote repositories
as a whole:

<center>
    <pre>[+][&lt;type>::][./]&lt;file path></pre>
		or
    <pre>[+][&lt;type>::]&lt;json repo spec></pre>
		or
    <pre>[+][&lt;type>::][&lt;scheme>://]&lt;domain>[:&lt;port>][/&lt;repository prefix>]</pre>
		or
	<pre>[+][&lt;type>::][&lt;scheme>://]&lt;host>[:&lt;port>][/&lt;repository prefix>]</pre>
</center>
` + topicocirefs.FileBasedUsage(),
	}
}
