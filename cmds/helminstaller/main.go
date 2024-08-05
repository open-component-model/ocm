package main

import (
	"os"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/helminstaller/app"
)

func main() {
	c := app.NewCliCommand(clictx.New(), nil)
	if err := c.Execute(); err != nil {
		os.Exit(1)
	}
}
