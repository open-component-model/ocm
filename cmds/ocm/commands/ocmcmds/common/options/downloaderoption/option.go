package downloaderoption

import (
	_ "github.com/open-component-model/ocm/pkg/contexts/ocm/download/handlers"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/optutils"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download"
	"github.com/open-component-model/ocm/pkg/listformat"
)

type Registration = optutils.Registration

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

func New(ctx ocm.Context) *Option {
	return &Option{optutils.NewRegistrationOption("downloader", "", "artifact downloader", Usage(ctx))}
}

type Option struct {
	optutils.RegistrationOption
}

func (o *Option) Register(ctx ocm.ContextProvider) error {
	for _, s := range o.Registrations {
		err := download.RegisterHandlerByName(ctx.OCMContext(), s.Name, s.Config,
			download.ForArtifactType(s.ArtifactType), download.ForMimeType(s.MediaType))
		if err != nil {
			return err
		}
	}
	return nil
}

func Usage(ctx ocm.Context) string {
	list := download.For(ctx).GetHandlers(ctx)
	return listformat.FormatListElements("", list) + `

See <CMD>ocm ocm-downloadhandlers</CMD> for further details on using
download handlers.
`
}
