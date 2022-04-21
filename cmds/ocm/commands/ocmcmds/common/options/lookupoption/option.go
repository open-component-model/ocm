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

package lookupoption

import (
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
)

func From(o *output.Options) *Option {
	var opt *Option
	o.Get(&opt)
	return opt
}

type Option struct {
	repooption.Option
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Spec, "lookup", "", "", "repository name or spec for closure lookup fallback")
}

func (o *Option) Usage() string {
	s := `
If a component lookup for building a reference closure is required
the <code>--lookup</code>  option can be used to specify a fallback
lookup repository. 
By default the component versions are searched in the repository
holding the component version for which the closure is determined.
For *Component Archives* this is never possible, because it only
contains a single component version. Therefore, in this scenario
this option must always be specified to be able to follow component
references.
`
	return s
}

func (o *Option) LookupComponentVersion(name string, vers string) (ocm.ComponentAccess, ocm.ComponentVersionAccess, error) {
	if o == nil {
		return nil, nil, nil
	}
	comp, err := o.Repository.LookupComponent(name)
	if err != nil {
		return nil, nil, err
	}
	v, err := comp.LookupVersion(vers)
	return comp, v, err
}
