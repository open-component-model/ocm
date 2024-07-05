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
OCM repositories can be configured to generate change events for
publish/subscribe systems, if there is a persistence provider
for the dedicated kind of OCM repository (for example OCI registry
based OCM repositories)

` + pubsub.PubSubUsage(attr.TypeScheme, attr.ProviderRegistry, true),
	}
}
