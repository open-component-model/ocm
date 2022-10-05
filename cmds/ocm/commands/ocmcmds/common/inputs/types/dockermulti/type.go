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

package dockermulti

import (
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
)

const TYPE = "dockermulti"

func init() {
	inputs.DefaultInputTypeScheme.Register(TYPE, inputs.NewInputType(TYPE, &Spec{}, usage))
}

const usage = `
This input type describes the composition of a multi-platform OCI image.
The various variants are taken from the local docker daemon. They should be 
built with the buildx command for cross platform docker builds.
The denoted images, as well as the wrapping image index is packed as OCI artefact set.

This blob type specification supports the following fields:
- **<code>variants</code>** *[]string*

  This REQUIRED property describes a set of  image names to import from the
  local docker daemon used to compose a resulting image index.

- **<code>repository</code>** *string*

  This OPTIONAL property can be used to specify the repository hint for the
  generated local artefact access. It is prefixed by the component name if
  it does not start with slash "/".`
