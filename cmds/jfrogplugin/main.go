package main

import (
	"fmt"
	"os"

	"ocm.software/ocm/api/ocm/plugin/ppi/cmds"
	jfrogppi "ocm.software/ocm/cmds/jfrogplugin/ppi"
)

func main() {
	plugin, err := jfrogppi.Plugin()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while creating plugin: %v\n", err)
		os.Exit(1)
	}
	if err := cmds.NewPluginCommand(plugin).Execute(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error while running plugin: %v\n", err)
		os.Exit(1)
	}
}
