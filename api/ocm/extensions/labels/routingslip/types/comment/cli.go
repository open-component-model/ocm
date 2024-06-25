package comment

import (
	"github.com/open-component-model/ocm/api/ocm/extensions/accessmethods/options"
	"github.com/open-component-model/ocm/api/utils/cobrautils/flagsets"
)

func ConfigHandler() flagsets.ConfigOptionTypeSetHandler {
	return flagsets.NewConfigOptionTypeSetHandler(
		Type, AddConfig,
		options.CommentOption,
	)
}

func AddConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	flagsets.AddFieldByOptionP(opts, options.CommentOption, config, "comment")
	return nil
}

var usage = `
An unstructured comment as entry in a routing slip.
`

var formatV1 = `
The type specific specification fields are:

- **<code>comment</code>**  *string*

  Any text as entry in a routing slip.
`
