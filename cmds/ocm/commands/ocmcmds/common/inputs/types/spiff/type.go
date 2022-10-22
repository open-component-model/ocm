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

package spiff

import (
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/file"
)

const TYPE = "spiff"

func init() {
	inputs.DefaultInputTypeScheme.Register(TYPE, inputs.NewInputType(TYPE, &Spec{}, usage(), ConfigHandler()))
}

func usage() string {
	return file.Usage("The path must denote a [spiff](https://github.com/mandelsoft/spiff) template relative the the resources file.") + `
- **<code>values</code>** *map[string]any*

  This OPTIONAL property describes an additioanl value binding for the template processing. It will be available
  under the node <code>values</code>.

- **<code>libraries</code>** *[]string*

  This OPTIONAL property describes a list of spiff libraries to include in template
  processing.
`
}
