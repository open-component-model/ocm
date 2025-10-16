package topicocmrefs

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	topicocirefs "ocm.software/ocm/cmds/ocm/topics/oci/refs"
)

func New(ctx clictx.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "ocm-references",
		Short: "notation for OCM references",
		Example: `
Complete Component Reference Specifications (including all optional arguments):

+ctf+directory::./ocm/ctf//ocm.software/ocmcli:0.7.0

oci::{"baseUrl":"ghcr.io","componentNameMapping":"urlPath","subPath":"open-component-model"}//ocm.software/ocmcli.0.7.0

oci::https://ghcr.io:443/open-component-model//ocm.software/ocmcli:0.7.0

oci::http://localhost:8080/local-component-repository//ocm.software/ocmcli:0.7.0

---

Short-Hand Component Reference Specifications (omitting optional arguments):

./ocm/ctf//ocm.software/ocmcli:0.7.0

ghcr.io/open-component-model//ocm.software/ocmcli:0.7.0

localhost:8080/local-component-repository//ocm.software/ocmcli:0.7.0 (defaulting to https)

http://localhost:8080/local-component-repository//ocm.software/ocmcli:0.7.0
`,
		Long: `
The command line client supports a special notation scheme for specifying
references to OCM components and repositories. This allows for specifying
references to any registry supported by the OCM toolset that can host OCM
components:

<center>
    <pre>[+][&lt;type>::][./]&lt;file path>//&lt;component id>[:&lt;version>]</pre>
</center>

or

<center>
	<pre>[+][&lt;type>::][&lt;json repo spec>//]&lt;component id>[:&lt;version>]</pre>
</center>

or

<center>
    <pre>[+][&lt;type>::][&lt;scheme>://]&lt;domain>[:&lt;port>][/&lt;repository prefix>]//&lt;component id>[:&lt;version]</pre>
</center>

or

<center>
	<pre>[+][&lt;type>::][&lt;scheme>://]&lt;host>[:&lt;port>][/&lt;repository prefix>]//&lt;component id>[:&lt;version]</pre>
</center>

---

Besides dedicated components it is also possible to denote repositories
as a whole:

<center>
    <pre>[+][&lt;type>::][./]&lt;file path></pre>
</center>

or

<center>
    <pre>[+][&lt;type>::]&lt;json repo spec></pre>
</center>

or

<center>
    <pre>[+][&lt;type>::][&lt;scheme>://]&lt;domain>[:&lt;port>][/&lt;repository prefix>]</pre>
</center>

or

<center>
	<pre>[+][&lt;type>::][&lt;scheme>://]&lt;host>[:&lt;port>][/&lt;repository prefix>]</pre>
</center>
` + topicocirefs.FileBasedUsage(),
	}
}
