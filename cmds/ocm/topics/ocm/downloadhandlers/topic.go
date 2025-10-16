package topicocmaccessmethods

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/downloaderoption"
)

func New(ctx clictx.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "ocm-downloadhandlers",
		Short: "List of all available download handlers",

		Long: `
A download handler can be used to process resources to be downloaded from
on OCM repository. By default, the blobs provided from the access method
(see <CMD>ocm ocm-accessmethods</CMD>) are used to store the resource content
in the local filesystem. Download handlers can be used to tweak this process.
They get access to the blob content and decide on their own what to do
with it, or how to transform it into files stored in the file system.

For example, a pre-registered helm download handler will store
OCI-based helm artifacts as regular helm archives in the local
file system.

### Handler Registration 

Programmatically any kind of handlers can be registered for various
download conditions. But this feature is available as command-line option, also.
New handlers can be provided by plugins. In general available handlers,
plugin-based or as part of the CLI coding are nameable using an hierarchical
namespace. Those names can be used by a <code>--downloader</code> option
to register handlers for various conditions for CLI commands like
<CMD>ocm download resources</CMD> (implicitly registered download handlers
can be enabled using the option <code>-d</code>).

Besides the activation constraints (resource type and media type of the
resource blob), it is possible to pass handler configuration controlling the
exact behaviour of the handler for selected artifacts.

The following handler names are possible:
` + downloaderoption.Usage(ctx.OCMContext()),
	}
}
