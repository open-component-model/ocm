package topicocmpubsub

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/pubsub"
)

func New(ctx clictx.Context) *cobra.Command {
	attr := pubsub.For(ctx)
	return &cobra.Command{
		Use:   "ocm-pubsub",
		Short: "List of all supported publish/subscribe implementations",

		Long: `
An OCM repository can be configured to propagate change events via a 
publish/subscribe system, if there is a persistence provider for the dedicated
repository type. If available any known publish/subscribe system can
be configured with <CMD>ocm set pubsub</CMD> and shown with
<CMD>ocm get pubsub</CMD>. Hereby, the pub/sub system 
is described by a typed specification.

` + pubsub.PubSubUsage(attr.TypeScheme, attr.ProviderRegistry, true),
	}
}
