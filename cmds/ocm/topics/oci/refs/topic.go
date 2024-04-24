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
+ctf+directory::./ocm/ctf//ocm.software/ocmcli/ocmcli-image:0.7.0@sha256:29c842be1ef1da67f6a1c07a3a3a8eb101bbcc4c80f174b87d147b341bca9625

oci::{"baseUrl": "ghcr.io"}//open-component-model/ocm/ocm.software/ocmcli/ocmcli-image:0.7.0@sha256:29c842be1ef1da67f6a1c07a3a3a8eb101bbcc4c80f174b87d147b341bca9625

oci::https://ghcr.io/open-component-model/ocm/ocm.software/ocmcli/ocmcli-image:0.7.0@sha256:29c842be1ef1da67f6a1c07a3a3a8eb101bbcc4c80f174b87d147b341bca9625
oci::https://ghcr.io//open-component-model/ocm/ocm.software/ocmcli/ocmcli-image:0.7.0@sha256:29c842be1ef1da67f6a1c07a3a3a8eb101bbcc4c80f174b87d147b341bca9625

oci::http://localhost:8080/ocm.software/ocmcli/ocmcli-image:0.7.0@sha256:29c842be1ef1da67f6a1c07a3a3a8eb101bbcc4c80f174b87d147b341bca9625
oci::http://localhost:8080//ocm.software/ocmcli/ocmcli-image:0.7.0@sha256:29c842be1ef1da67f6a1c07a3a3a8eb101bbcc4c80f174b87d147b341bca9625

ubuntu:24.04
ubuntu

tensorflow/tensorflow:2.15.0
tensorflow/tensorflow
`,
		Long: `
The command line client supports a special notation scheme for specifying
references to instances of oci like registries. This allows for specifying
references to any registry supported by the OCM toolset that can host OCI
artifacts. As a subset the regular OCI artifact notation used for docker
images are possible:

<center>
    <pre>[+][&lt;type>::][./][&lt;file path>//&lt;repository>[:&lt;tag>][@&lt;digest>]</pre>
</center>

or  

<center>
    <pre>[+][&lt;type>::][&lt;json repo spec>//]&lt;repository>[:&lt;tag>][@&lt;digest>]</pre>
</center>

Notice that if you specify the &lt;type> in the beginning of this
notation AND in the &lt;json repo spec>, the types have to match
(but there is no reason to specify the type in both places).

or

<center>
	<pre>[+][&lt;type>::][&lt;scheme>://]&lt;domain>[:&lt;port>][/]/&lt;repository>[:&lt;tag>][@&lt;digest>]</pre>
</center>

Notice that this notation optionally also allows a double slash to
seperate &lt;domain>[:&lt;port>] and &lt;repository>. While it is
not necessary for unambiguous parsing here, it is supported for
consistency with the other notations.

or  

<center>
	<pre>[+][&lt;type>::][&lt;scheme>://]&lt;host>:&lt;port>/&lt;repository>[:&lt;tag>][@&lt;digest>]</pre>
</center>

Notice that &lt;port> is required in this notation. Without &lt;port>,
this notation would be ambiguous with the docker library notation 
mentioned below.

or

<center>
	<pre>[+][&lt;type>::][&lt;scheme>://]&lt;host>[:&lt;port>]//&lt;repository>[:&lt;tag>][@&lt;digest>]</pre>
</center>

Notice the double slash (//) before the &lt;repository>. This serves as
a clear separator between &lt;host>[:&lt;port>] and &lt;repository>.
Thus, with this notation, the port is optional and can therefore be
omitted without creating ambiguity with the docker library notation
mentioned below.

or

<center>
    <pre>&lt;docker library>[:&lt;tag>][@&lt;digest>]</pre>
</center>

or  

<center>
    <pre>&lt;docker repository>/&lt;docker image>[:&lt;tag>][@&lt;digest>]</pre>
</center>

---

Besides dedicated artifacts it is also possible to denote registries
as a whole:

<center>
	<pre>[+][&lt;type>::][./]&lt;file path></pre>
</center>

or  

<center>
	<pre>[+][&lt;type>::]&lt;json repo spec></pre>
</center>

Notice that if you specify the &lt;type> in the beginning of this
notation AND in the &lt;json repo spec>, the types have to match
(but there is no reason to specify the type in both places).

or  

<center>
    <pre>[+][&lt;type>::][&lt;scheme>://]&lt;domain>[:&lt;port>]</pre>
</center>

or  

<center>
	<pre>[+][&lt;type>::][&lt;scheme>://]&lt;host>[:&lt;port>]</pre>
</center>

Notice that &lt;port> is optional in this notation since this cannot be
an image reference and therefore cannot be ambiguous with the docker
library notation.
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
