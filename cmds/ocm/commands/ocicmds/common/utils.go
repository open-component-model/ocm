package common

import (
	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
)

type OptionCompleter interface {
	CompleteWithSession(ctx clictx.OCI, session oci.Session) error
}

func CompleteOptionsWithContext(ctx clictx.Context, session oci.Session) options.OptionsProcessor {
	return func(opt options.Options) error {
		if c, ok := opt.(OptionCompleter); ok {
			return c.CompleteWithSession(ctx.OCI(), session)
		}
		if c, ok := opt.(options.OptionWithCLIContextCompleter); ok {
			return c.Configure(ctx)
		}
		if c, ok := opt.(options.SimpleOptionCompleter); ok {
			return c.Complete()
		}
		return nil
	}
}
