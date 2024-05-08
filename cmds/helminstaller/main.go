package main

import (
	"os"

	"github.com/open-component-model/ocm/cmds/helminstaller/app"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

func main() {
	c := app.NewCliCommand(clictx.New(), nil)
	if err := c.Execute(); err != nil {
		os.Exit(1)
	}
}
