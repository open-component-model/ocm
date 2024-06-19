package main

import (
	"os"

	"github.com/open-component-model/ocm/cmds/subcmdplugin/cmds/group"
	// enable mandelsoft plugin logging configuration.
	_ "github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/logging"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/clicmd"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds"
	"github.com/open-component-model/ocm/pkg/version"
)

func main() {
	p := ppi.NewPlugin("cliplugin", version.Get().String())

	p.SetShort("Demo plugin with a simple cli extension")
	p.SetLong("The plugin offers the top-level command group with sub command demo")

	cmd, err := clicmd.NewCLICommand(group.New())
	if err != nil {
		os.Exit(1)
	}
	p.RegisterCommand(cmd)

	// fmt.Printf("CMD ARGS: %v\n", os.Args[1:])
	err = cmds.NewPluginCommand(p).Execute(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}
}
