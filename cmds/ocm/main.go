// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/open-component-model/ocm/cmds/ocm/app"
	cmd "github.com/open-component-model/ocm/cmds/ocm/clictx"
)

func main() {
	c := app.NewCliCommand(cmd.DefaultContext())

	if err := c.Execute(); err != nil {
		os.Exit(1)
	}
}
