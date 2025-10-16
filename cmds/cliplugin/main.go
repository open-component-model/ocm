package main

import (
	"os"

	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/ocm/plugin/ppi/clicmd"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds"
	// enable mandelsoft plugin logging configuration.
	_ "ocm.software/ocm/api/ocm/plugin/ppi/logging"
	"ocm.software/ocm/api/version"
	"ocm.software/ocm/cmds/cliplugin/cmds/check"
)

func main() {
	p := ppi.NewPlugin("cliplugin", version.Get().String())

	p.SetShort("Demo plugin with a simple cli extension")
	p.SetLong("The plugin offers the check command for object type rhubarb to check the rhubarb season.")

	cmd, err := clicmd.NewCLICommand(check.New(), clicmd.WithCLIConfig(), clicmd.WithObjectType("rhubarb"), clicmd.WithVerb("check"))
	if err != nil {
		os.Exit(1)
	}
	p.RegisterCommand(cmd)
	p.ForwardLogging()

	p.RegisterConfigType(check.RhabarberType)
	p.RegisterConfigType(check.RhabarberTypeV1)
	err = cmds.NewPluginCommand(p).Execute(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}
}
