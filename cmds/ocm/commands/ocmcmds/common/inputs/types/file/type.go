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

package file

import (
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/pkg/mime"
)

const TYPE = "file"

func init() {
	inputs.DefaultInputTypeScheme.Register(TYPE, inputs.NewInputType(TYPE, &Spec{}, usage))
}

const usage = `
- Input type <code>file</code>

  The path must denote a file relative the the resources file.
  The content is compressed if the <code>compress</code> field
  is set to <code>true</code>.

  This blob type specification supports the following fields: 
  - **<code>path</code>** *string*

    This REQUIRED property describes the file path to the helm chart relative to the
    resource file location.

  - **<code>mediaType</code>** *string*

    This OPTIONAL property describes the media type to store with the local blob.
    The default media type is ` + mime.MIME_OCTET + ` and
    ` + mime.MIME_GZIP + ` if compression is enabled.

  - **<code>compress</code>** *bool*

    This OPTIONAL property describes whether the file content should be stored
    compressed or not.`
