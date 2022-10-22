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

package helm

import (
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
)

const TYPE = "helm"

func init() {
	inputs.DefaultInputTypeScheme.Register(TYPE, inputs.NewInputType(TYPE, &Spec{}, usage, ConfigHandler()))
}

const usage = `
The path must denote an helm chart archive or directory
relative to the resources file.
The denoted chart is packed as an OCI artefact set.
Additional provider info is taken from a file with the same name
and the suffix <code>.prov</code>.

If the chart should just be stored as archive, please use the 
type <code>file</code> or <code>dir</code>.

This blob type specification supports the following fields: 
- **<code>path</code>** *string*

  This REQUIRED property describes the file path to the helm chart relative to the
  resource file location.

- **<code>version</code>** *string*

  This OPTIONAL property can be set to configure an explicit version hint.
  If not specified the versio from the chart will be used.
  Basically, it is a good practice to use the component version for local resources
  This can be achieved by using templating for this attribute in the resource file.
`
