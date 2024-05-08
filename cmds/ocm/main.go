package main

import (
	"fmt"
	"os"

	"github.com/open-component-model/ocm/cmds/ocm/app"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

func main() {
	c, err := app.NewCliCommandForArgs(clictx.DefaultContext(), os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
	if err := c.Execute(); err != nil {
		os.Exit(1)
	}
}
