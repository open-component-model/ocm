package main

import (
	"os"

	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/ocm/plugin/ppi/clicmd"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds"
	// enable mandelsoft plugin logging configuration.
	_ "ocm.software/ocm/api/ocm/plugin/ppi/logging"
	"ocm.software/ocm/api/version"
	"ocm.software/ocm/cmds/subcmdplugin/cmds/group"
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
