package main

import (
	"os"

	// enable mandelsoft plugin logging configuration.
	_ "github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/logging"

	"github.com/open-component-model/ocm/cmds/cliplugin/cmds/check"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/clicmd"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds"
	"github.com/open-component-model/ocm/pkg/version"
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
