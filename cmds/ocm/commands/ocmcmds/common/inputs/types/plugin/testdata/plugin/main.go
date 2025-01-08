package main

import (
	"os"

	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/plugin/testdata/plugin/app"
)

func main() {
	err := app.Run(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}
}
