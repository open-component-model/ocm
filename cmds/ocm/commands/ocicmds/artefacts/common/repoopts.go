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

package common

import (
	"github.com/gardener/ocm/cmds/ocm/clictx"
	"github.com/gardener/ocm/pkg/oci"
	"github.com/gardener/ocm/pkg/runtime"

	"github.com/spf13/pflag"
)

type RepositoryOptions struct {
	Repository oci.Repository
	Spec       string
}

func (o *RepositoryOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Spec, "repo", "r", "", "repository name or spec")
}

func (o *RepositoryOptions) Complete(ctx clictx.Context) error {
	var err error
	if o.Spec != "" {
		o.Repository, err = ctx.OCI().DetermineRepository(o.Spec)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *RepositoryOptions) Usage() string {

	s := `
If the repository/registry option is specified, the given names are interpreted
relative to the specified registry using the syntax

<center><code>&lt;OCI repository name>[:&lt;tag>][@&lt;digest>]</code></center>

If no <code>--repo</code> option is specified the given names are interpreted 
as extended CI artefact references.

<center><code>[&lt;repo type>::]&lt;host>[:&lt;port>]/&lt;OCI repository name>[:&lt;tag>][@&lt;digest>]</code></center>

The <code>--repo</code> option takes a repository/OCI registry specification:

<center><code>[&lt;repo type>::]&lt;configured name>|&lt;file path>|&lt;spec json></code></center>

For the *Common Transport Format* the types <code>directory</code>,
<code>tar</code> or <code>tgz</code> are possible.

Using the JSON variant any repository type supported by the 
linked library can be used:
`
	types := runtime.KindNames(oci.DefaultContext().RepositoryTypes())
	for _, t := range types {
		s += "- `" + t + "`\n"
	}
	return s
}
