package main

import (
	"os"

	"github.com/open-component-model/ocm/cmds/cliplugin/cmds/rhabarber"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/clicmd"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds"
	"github.com/open-component-model/ocm/pkg/version"
)

func main() {
	p := ppi.NewPlugin("cliplugin", version.Get().String())

	p.SetShort("Demo plugin with a simple cli extension")
	p.SetLong("The plugin offers the top-level command rhabarber")
	cmd, err := clicmd.NewCLICommand(rhabarber.New(), clicmd.WithCLIConfig())
	if err != nil {
		os.Exit(1)
	}
	p.RegisterCommand(cmd)
	p.RegisterConfigType(rhabarber.RhabarberType)
	p.RegisterConfigType(rhabarber.RhabarberTypeV1)
	err = cmds.NewPluginCommand(p).Execute(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}
}
