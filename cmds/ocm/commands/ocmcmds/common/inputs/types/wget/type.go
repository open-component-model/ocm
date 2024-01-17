// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package directory

import (
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/pkg/mime"
)

const TYPE = "wget"

func init() {
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(TYPE, &Spec{}, usage, ConfigHandler()))
}

const usage = `
The <code>url</code> is the url pointing to the http endpoint from which a resource is 
downloaded. The <code>mimeType</code> can be used to specify the MIME type of the 
resource.

This blob type specification supports the following fields:
- **<code>url</code>** *string*

  This REQUIRED property describes the url from which the resource is to be
  downloaded.

- **<code>mediaType</code> *string*

  This OPTIONAL property describes the media type of the resource to be 
  downloaded. The default media type is ` + mime.MIME_OCTET + `.
`
