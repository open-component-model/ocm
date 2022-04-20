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

package get

import (
	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/spf13/pflag"
)

func AttachedFrom(o options.OptionSetProvider) *Attached {
	var opt *Attached
	o.AsOptionSet().Get(&opt)
	return opt
}

type Attached struct {
	Flag bool
}

var _ options.Condition = (*Attached)(nil)
var _ options.Options = (*Attached)(nil)

func (a *Attached) IsTrue() bool {
	return a.Flag
}

func (a *Attached) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&a.Flag, "attached", "a", false, "show attached artefacts")
}
