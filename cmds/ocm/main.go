package main

import (
	"fmt"
	"os"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/ocm/app"
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
