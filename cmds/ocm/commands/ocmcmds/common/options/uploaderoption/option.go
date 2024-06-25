package uploaderoption

import (
	"github.com/open-component-model/ocm/api/ocm"
	"github.com/open-component-model/ocm/api/ocm/extensions/blobhandler"
	"github.com/open-component-model/ocm/api/utils/listformat"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/optutils"
	"github.com/open-component-model/ocm/cmds/ocm/common/options"
)

type Registration = optutils.Registration

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

func New(ctx ocm.Context) *Option {
	return &Option{optutils.NewRegistrationOption("uploader", "", "repository uploader", Usage(ctx))}
}

type Option struct {
	optutils.RegistrationOption
}

func (o *Option) Register(ctx ocm.ContextProvider) error {
	for _, s := range o.Registrations {
		err := blobhandler.RegisterHandlerByName(ctx.OCMContext(), s.Name, s.Config,
			blobhandler.ForArtifactType(s.ArtifactType), blobhandler.ForMimeType(s.MediaType))
		if err != nil {
			return err
		}
	}
	return nil
}

func Usage(ctx ocm.Context) string {
	list := blobhandler.For(ctx).GetHandlers(ctx)
	return listformat.FormatListElements("", list) + `

See <CMD>ocm ocm-uploadhandlers</CMD> for further details on using
upload handlers.
`
}
