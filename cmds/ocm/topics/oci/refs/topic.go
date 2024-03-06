// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package topicocirefs

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf"
)

func New(ctx clictx.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "oci-references",
		Short: "notation for OCI references",
		Example: `
ghcr.io/mandelsoft/cnudie:1.0.0
`,
		Long: `
The command line client supports a special notation scheme for specifying
references to instances of oci like registries. This allows for specifying
references to any registry supported by the OCM toolset that can host OCI
artifacts. As a subset the regular OCI artifact notation used for docker
images are possible:

<center>
    <pre>[+][&lt;type>::][./][&lt;file path>//&lt;repository>[:&lt;tag>][@&lt;digest>]</pre>
        or
	<pre>[+][&lt;type>::][&lt;scheme>:://]&lt;domain>[:&lt;port>]/&lt;repository>[:&lt;tag>][@&lt;digest>]</pre>
        or
    <pre>[+][&lt;type>::][&lt;json repo spec>//]&lt;repository>[:&lt;tag>][@&lt;digest>]</pre>
        or
	<pre>[+][&lt;type>::][&lt;scheme>://]&lt;host>:<port>[:&lt;tag>][@&lt;digest>]</pre>
		Notice that <port> is required in this notation. Without <port>, this
		notation would be ambiguous with the docker library notation mentioned
		below.  
		or
    <pre>&lt;docker library>[:&lt;tag>][@&lt;digest>]</pre>
        or
    <pre>&lt;docker repository>/&lt;docker image>[:&lt;tag>][@&lt;digest>]</pre>
</center>

Besides dedicated artifacts it is also possible to denote registries
as a whole:

<center>
    <pre>[+][&lt;type>::][&lt;scheme>:://]&lt;domain>[:&lt;port>]</pre>
        or
    <pre>[+][&lt;type>::]&lt;json repo spec></pre>
        or
    <pre>[+][&lt;type>::][./]&lt;file path></pre>
</center>
` + FileBasedUsage(),
	}
}

func FileBasedUsage() string {
	s := `
The optional <code>+</code> is used for file based implementations
(Common Transport Format) to indicate the creation of a not yet existing
file.

The **type** may contain a file format qualifier separated by a <code>+</code>
character. The following formats are supported: `

	list := ""
	for _, f := range ctf.SupportedFormats() {
		list = fmt.Sprintf("%s, <code>%s</code>", list, f)
	}
	return s + list[2:]
}
