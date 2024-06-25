package main

import (
	"os"

	"github.com/open-component-model/ocm/api/clictx"
	"github.com/open-component-model/ocm/cmds/helminstaller/app"
)

func main() {
	c := app.NewCliCommand(clictx.New(), nil)
	if err := c.Execute(); err != nil {
		os.Exit(1)
	}
}
