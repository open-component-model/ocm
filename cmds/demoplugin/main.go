package main

import (
	"os"

	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds"
	"ocm.software/ocm/api/version"
	"ocm.software/ocm/cmds/demoplugin/accessmethods"
	"ocm.software/ocm/cmds/demoplugin/config"
	"ocm.software/ocm/cmds/demoplugin/uploaders"
	"ocm.software/ocm/cmds/demoplugin/valuesets"
)

func main() {
	p := ppi.NewPlugin("demo", version.Get().String())

	p.SetShort("demo plugin")
	p.SetLong("plugin providing access to temp files and a check routing slip entry.")
	p.SetConfigParser(config.GetConfig)

	p.RegisterAccessMethod(accessmethods.New())
	u := uploaders.New()
	p.RegisterUploader("testArtifact", "", u)
	p.RegisterValueSet(valuesets.New())
	err := cmds.NewPluginCommand(p).Execute(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}
}
