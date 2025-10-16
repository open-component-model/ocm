package app

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm/tools/toi/support"
	"ocm.software/ocm/cmds/helminstaller/app/driver"
	"ocm.software/ocm/cmds/helminstaller/app/driver/helm"
)

func NewCliCommand(ctx clictx.Context, d driver.Driver) *cobra.Command {
	if d == nil {
		d = helm.New()
	}
	return support.NewCLICommand(ctx.OCMContext(), "helmbootstrapper", New(d))
}
