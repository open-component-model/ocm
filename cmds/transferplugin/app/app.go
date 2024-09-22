package app

import (
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds"
	"ocm.software/ocm/api/version"
	"ocm.software/ocm/cmds/transferplugin/config"
	"ocm.software/ocm/cmds/transferplugin/transferhandlers"
)

func Run(args []string, opts ...cmds.Option) error {
	p := ppi.NewPlugin("transferplugin", version.Get().String())

	p.SetShort("demo transfer handler plugin")
	p.SetLong("plugin providing a transfer handler to enable value transport for dedicated external repositories.")
	p.SetConfigParser(config.GetConfig)

	err := p.RegisterTransferHandler(transferhandlers.New())
	if err != nil {
		return err
	}
	return cmds.NewPluginCommand(p, opts...).Execute(args)
}
