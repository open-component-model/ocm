// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"github.com/gardener/ocm/cmds/ocm/app"
	cmd "github.com/gardener/ocm/cmds/ocm/clictx"
)

func main() {
	c := app.NewCliCommand(cmd.DefaultContext())

	if err := c.Execute(); err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}
