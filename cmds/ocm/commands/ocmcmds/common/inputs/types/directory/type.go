// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package directory

import (
	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/v2/pkg/mime"
)

const TYPE = "dir"

func init() {
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(TYPE, &Spec{}, usage, ConfigHandler()))
}

const usage = `
The path must denote a directory relative to the resources file, which is packed
with tar and optionally compressed
if the <code>compress</code> field is set to <code>true</code>. If the field
<code>preserveDir</code> is set to true the directory itself is added to the tar.
If the field <code>followSymLinks</code> is set to <code>true</code>, symbolic
links are not packed but their targets files or folders.
With the list fields <code>includeFiles</code> and <code>excludeFiles</code> it is 
possible to specify which files should be included or excluded. The values are
regular expression used to match relative file paths. If no includes are specified
all file not explicitly excluded are used.

This blob type specification supports the following fields: 
- **<code>path</code>** *string*

  This REQUIRED property describes the file path to directory relative to the
  resource file location.

- **<code>mediaType</code>** *string*

  This OPTIONAL property describes the media type to store with the local blob.
  The default media type is ` + mime.MIME_TAR + ` and
  ` + mime.MIME_GZIP + ` if compression is enabled.

- **<code>compress</code>** *bool*

  This OPTIONAL property describes whether the file content should be stored
  compressed or not.

- **<code>preserveDir</code>** *bool*

  This OPTIONAL property describes whether the specified directory with its
  basename should be included as top level folder.

- **<code>followSymlinks</code>** *bool*

  This OPTIONAL property describes whether symbolic links should be followed or
  included as links.

- **<code>excludeFiles</code>** *list of regex*

  This OPTIONAL property describes regular expressions used to match files 
  that should NOT be included in the tar file. It takes precedence over
  the include match.

- **<code>includeFiles</code>** *list of regex*

  This OPTIONAL property describes regular expressions used to match files 
  that should be included in the tar file. If this option is not given
  all files not explicitly excluded are used.
`
