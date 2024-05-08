package verify

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/cmds/signing"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

var (
	Names = names.Components
	Verb  = verbs.Verify
)

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return signing.NewCommand(ctx, "Verify signature of", false,
		[]string{"verified", "verifying signature of"},
		desc, "$ ocm verify componentversion --signature mandelsoft --public-key=mandelsoft.key ghcr.io/mandelsoft/kubelink",
		utils.Names(Names, names...)...)
}

var desc = `
If no signature name is given, only the digests are validated against the
registered ones at the component version.
`
