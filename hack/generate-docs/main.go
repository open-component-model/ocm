// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"github.com/open-component-model/ocm/cmds/ocm/app"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

func main() {
	//nolint: forbidigo // Intentional Println because this is a supplementary tool.
	fmt.Println("> Generate Docs for OCM CLI")

	if len(os.Args) != 2 { // expect 2 as the first one is the program name
		fmt.Fprintf(os.Stderr, "Expected exactly one argument, but got %d", len(os.Args)-1)
		os.Exit(1)
	}
	outputDir := os.Args[1]

	// clear generated
	check(os.RemoveAll(outputDir))
	check(os.MkdirAll(outputDir, os.ModePerm))

	cmd := app.NewCliCommand(clictx.DefaultContext())
	cmd.DisableAutoGenTag = true
	check(GenMarkdownTree(cmd, outputDir))
	//nolint: forbidigo // Intentional Println because this is a supplementary tool.
	fmt.Printf("Successfully written docs to %s\n", outputDir)
}

func check(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
}
