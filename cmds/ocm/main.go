// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"github.com/open-component-model/ocm/v2/cmds/ocm/app"
	"github.com/open-component-model/ocm/v2/pkg/contexts/clictx"
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
