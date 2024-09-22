package main

import (
	"os"

	"ocm.software/ocm/cmds/transferplugin/app"
)

func main() {
	err := app.Run(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}
}
