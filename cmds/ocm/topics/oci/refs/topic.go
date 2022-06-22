// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package topicocirefs

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/clictx"
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
artefacts. As a subset the regular OCI artefact notation used for docker
images are possible:

<center>
    <pre>[+][&lt;type>::][./][&lt;file path>//&lt;repository>[:&lt;tag>][@&lt;digest>]</pre>
        or
    <pre>[&lt;type>::][&lt;json repo spec>//]&lt;repository>[:&lt;tag>][@&lt;digest>]</pre>
        or
    <pre>[&lt;type>::][&lt;scheme>:://]&lt;domain>[:&lt;port>/]&lt;repository>[:&lt;tag>][@&lt;digest>]</pre>
        or
    <pre>&lt;docker library>[:&lt;tag>][@&lt;digest>]</pre>
        or
    <pre>&lt;docker repository>/&lt;docker image>[:&lt;tag>][@&lt;digest>]</pre>
</center>

Besides dedicated artefacts it is also possible to denote registries
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
