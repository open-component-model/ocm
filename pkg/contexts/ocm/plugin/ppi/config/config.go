package config

import (
	"context"

	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/command"
	"github.com/open-component-model/ocm/pkg/runtime"
)

func init() {
	command.RegisterCommandConfigHandler(&commandHandler{})
}

type commandHandler struct{}

func (c commandHandler) HandleConfig(ctx context.Context, data []byte) (context.Context, error) {
	var err error

	octx := ocm.DefaultContext()
	ctx = octx.BindTo(ctx)
	if len(data) != 0 {
		_, err = octx.ConfigContext().ApplyData(data, runtime.DefaultYAMLEncoding, " cli config")
		// Ugly, enforce configuration update
		octx.GetResolver()
	}
	return ctx, err
}
