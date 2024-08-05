package verify

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/cmds/signing"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/utils"
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
