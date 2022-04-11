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

package comppathopt

import (
	"github.com/gardener/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/gardener/ocm/cmds/ocm/pkg/output"
	metav1 "github.com/gardener/ocm/pkg/ocm/compdesc/meta/v1"

	"github.com/spf13/pflag"
)

func From(o *output.Options) *Option {
	var opt *Option
	o.Get(&opt)
	return opt
}

type Option struct {
	Active bool
	Ids    []metav1.Identity
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&o.Active, "path", "p", false, "follow component references")
}

// Complete consumes path identities if option is activated
func (o *Option) Complete(args []string) ([]string, error) {
	var err error
	rest := args
	if o.Active {
		o.Ids, rest, err = common.ConsumeIdentities(args, ";")
	}
	return rest, err
}

func (o *Option) Usage() string {

	s := `
The <code>--path</code> options accets a sequence of identities,
that will be used to follow component references a the specified
component(s).

In identity is given by a sequence of arguments starting with a
plain name value argument followed by any number of attribute assignments
of the form <code>&lt;<name>=&lt;value></code>.
The identity sequence stops at the end of the command line or with a sole
<code>;</code> argument, if other arguments are required for further purposes.
`
	return s
}
