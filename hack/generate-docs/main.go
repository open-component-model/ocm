// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"github.com/open-component-model/ocm/v2/cmds/ocm/app"
	"github.com/open-component-model/ocm/v2/hack/generate-docs/cobradoc"
	"github.com/open-component-model/ocm/v2/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/attrs/plugindirattr"
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
