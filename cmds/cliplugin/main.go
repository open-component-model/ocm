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

	p.RegisterCommand(clicmd.NewCLICommand(rhabarber.New()))
	err := cmds.NewPluginCommand(p).Execute(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}
}
