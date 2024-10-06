package app

import (
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds"
	"ocm.software/ocm/api/version"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/plugin/testdata/plugin/inputhandlers"
)

func New() (ppi.Plugin, error) {
	p := ppi.NewPlugin("input", version.Get().String())

	p.SetShort("fake input plugin")
	p.SetLong("providing fake input")

	err := p.RegisterInputType(inputhandlers.New())
	if err != nil {
		return nil, err
	}
	return p, nil
}

func Run(args []string, opts ...cmds.Option) error {
	p, err := New()
	if err != nil {
		return err
	}
	return cmds.NewPluginCommand(p, opts...).Execute(args)
}
