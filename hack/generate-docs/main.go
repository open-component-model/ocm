// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"github.com/open-component-model/ocm/cmds/ocm/app"
	"github.com/open-component-model/ocm/hack/generate-docs/cobradoc"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

func main() {
	//nolint: forbidigo // Intentional Println because this is a supplementary tool.
	fmt.Println("> Generate Docs for OCM CLI")

	if len(os.Args) != 2 { // expect 2 as the first one is the program name
		fmt.Fprintf(os.Stderr, "Expected exactly one argument, but got %d", len(os.Args)-1)
		os.Exit(1)
	}

	cmd := app.NewCliCommand(clictx.DefaultContext())
	cmd.DisableAutoGenTag = true
	cobradoc.Generate("OCM CLI", cmd, os.Args[1], true)
}
