// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"github.com/open-component-model/ocm/hack/generate-docs/cobradoc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds"
	"github.com/open-component-model/ocm/pkg/version"
)

func main() {
	fmt.Println("> Generate Docs for OCM Plugins")

	if len(os.Args) != 2 { // expect 2 as the first one is the program name
		fmt.Fprintf(os.Stderr, "Expected exactly one argument, but got %d", len(os.Args)-1)
		os.Exit(1)
	}

	p := ppi.NewPlugin("plugin", version.Get().String())
	p.SetLong(cmds.Description(p.Name()))
	p.SetShort("OCM Plugin")
	cmd := cmds.NewPluginCommand(p).Command()
	cmd.DisableAutoGenTag = true
	cobradoc.Generate("OCM Plugin", cmd, os.Args[1], true)
}
