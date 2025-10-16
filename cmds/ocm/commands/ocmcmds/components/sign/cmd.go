package sign

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
	Verb  = verbs.Sign
)

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return signing.NewCommand(ctx, "Sign", true,
		[]string{"signed", "signing"},
		desc, "$ ocm sign componentversion --signature mysignature --private-key=my.key ghcr.io/open-component-model/ocm//ocm.software/ocmcli:0.17.0",
		utils.Names(Names, names...)...)
}

var desc = ``
