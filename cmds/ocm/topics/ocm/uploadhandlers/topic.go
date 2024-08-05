package topicocmaccessmethods

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/uploaderoption"
)

func New(ctx clictx.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "ocm-uploadhandlers",
		Short: "List of all available upload handlers",

		Long: `
An upload handler is used to process resources using the access method
<code>localBlob</code> transferred into an OCM
repository. They may decide to store the content in some other 
storage repository. This may be an additional storage location or it
may replace the storage of the resource as local blob.
If an additional storage location is chosen, the local access method
is kept and the additional location can be registered in the component
descriptor as <code>globalAccess</code> attribute of the local access
specification.

For example, there is a default upload handler responsible for OCI artifact
blobs, which provides regular OCI artifacts for a local blob, if
the target OCM repository is based on an OCI registry. Hereby, the
<code>referenceName</code> attribute will be used to calculate a
meaningful OCI repository name based on the repository prefix
of the OCM repository (parallel to <code>component-descriptors</code> prefix
used to store the component descriptor artifacts).

### Handler Registration 

Programmatically any kind of handlers can be registered for various
upload conditions. But this feature is available as command-line option, also.
New handlers can be provided by plugins. In general available handlers,
plugin-based or as part of the CLI coding are nameable using an hierarchical
namespace. Those names can be used by a <code>--uploader</code> option
to register handlers for various conditions for CLI commands like
<CMD>ocm transfer componentversions</CMD> or <CMD>ocm transfer commontransportarchive</CMD>.

Besides the activation constraints (resource type and media type of the
resource blob), it is possible to pass a target configuration controlling the
exact behaviour of the handler for selected artifacts.

The following handler names are possible:
` + uploaderoption.Usage(ctx.OCMContext()),
	}
}
