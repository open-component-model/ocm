package main

import (
	"fmt"
	"os"

	"ocm.software/ocm/cmds/ocm/app"
	"ocm.software/ocm/hack/generate-docs/cobradoc"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm/extensions/attrs/plugindirattr"
)

func main() {
	//nolint: forbidigo // Intentional Println because this is a supplementary tool.
	fmt.Println("> Generate Docs for OCM CLI")

	if len(os.Args) != 2 { // expect 2 as the first one is the program name
		fmt.Fprintf(os.Stderr, "Expected exactly one argument, but got %d", len(os.Args)-1)
		os.Exit(1)
	}

	ctx := clictx.DefaultContext()
	plugindirattr.Set(ctx.AttributesContext(), "")
	cmd := app.NewCliCommand(ctx)
	cmd.DisableAutoGenTag = true
	cobradoc.Generate("OCM CLI", cmd, os.Args[1], true)
}
