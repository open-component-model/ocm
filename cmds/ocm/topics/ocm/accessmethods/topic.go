package topicocmaccessmethods

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm"
)

func New(ctx clictx.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "ocm-accessmethods",
		Short: "List of all supported access methods",

		Long: `
Access methods are used to handle the access to the content of artifacts
described in a component version. Therefore, an artifact entry contains
an access specification describing the access attributes for the dedicated
artifact.

` + ocm.AccessUsage(ctx.OCMContext().AccessMethods(), true),
	}
}
