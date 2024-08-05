package main

import (
	"fmt"
	"os"

	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds"
	"ocm.software/ocm/api/version"
	"ocm.software/ocm/hack/generate-docs/cobradoc"
)

func main() {
	fmt.Println("> Generate Docs for OCM Plugins")

	if len(os.Args) != 2 { // expect 2 as the first one is the program name
		fmt.Fprintf(os.Stderr, "Expected exactly one argument, but got %d", len(os.Args)-1)
		os.Exit(1)
	}

	p := ppi.NewPlugin("plugin", version.Get().String())
	p.SetLong(cmds.Description(p.Name()))
	p.SetShort("OCM Plugin")
	cmd := cmds.NewPluginCommand(p).Command()
	cmd.DisableAutoGenTag = true
	cobradoc.Generate("OCM Plugin", cmd, os.Args[1], true)
}
