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

package topicocmrefs

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/pkg/contexts/clictx"

	topicocirefs "github.com/open-component-model/ocm/cmds/ocm/topics/oci/refs"
)

func New(ctx clictx.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "ocm-references",
		Short: "notation for OCM references",
		Example: `
ghcr.io/mandelsoft/cnudie//github.com/mandelsoft/pause:1.0.0

ctf+tgz::./ctf
`,
		Long: `
The command line client supports a special notation scheme for specifying
references to OCM components and repositories. This allows for specifying
references to any registry supported by the OCM toolset that can host OCM
components:

<center>
    <pre>[+][&lt;type>::][./][&lt;file path>//&lt;component id>[:&lt;version>]</pre>
        or
    <pre>[+][&lt;type>::]&lt;domain>[:&lt;port>][/&lt;repository prefix>]//&lt;component id>[:&lt;version]</pre>
        or
    <pre>[&lt;type>::][&lt;json repo spec>//]&lt;component id>[:&lt;version>]</pre>

</center>

Besides dedicated components it is also possible to denote repositories
as a whole:

<center>
    <pre>[+][&lt;type>::][&lt;scheme>:://]&lt;domain>[:&lt;port>][/&lt;repository prefix>]</pre>
        or
    <pre>[+][&lt;type>::]&lt;json repo spec></pre>
        or
    <pre>[+][&lt;type>::][./]&lt;file path></pre>
</center>
` + topicocirefs.FileBasedUsage(),
	}
}
