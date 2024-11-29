package main

import (
	"os"

	"ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds"
	"ocm.software/ocm/api/version"
	"ocm.software/ocm/cmds/jfrogplugin/uploaders"
)

func main() {
	p := ppi.NewPlugin("jfrog", version.Get().String())

	p.SetShort("JFrog plugin")
	p.SetLong("plugin providing custom functions related to interacting with JFrog Repositories (e.g. Artifactory).")
	p.SetConfigParser(uploaders.GetConfig)

	u := uploaders.New()
	if err := p.RegisterUploader(artifacttypes.HELM_CHART, "", u); err != nil {
		panic(err)
	}
	err := cmds.NewPluginCommand(p).Execute(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}
}
