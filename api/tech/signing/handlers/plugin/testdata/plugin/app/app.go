package app

import (
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds"
	"ocm.software/ocm/api/tech/signing/handlers/plugin/testdata/plugin/signinghandlers"
	"ocm.software/ocm/api/version"
)

func New() (ppi.Plugin, error) {
	p := ppi.NewPlugin("signing", version.Get().String())

	p.SetShort("fake signing plugin")
	p.SetLong("providing fake signing and verification")

	err := p.RegisterSigningHandler(signinghandlers.New())
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
