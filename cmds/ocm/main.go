// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/open-component-model/ocm/cmds/ocm/app"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

func main() {
	c := app.NewCliCommand(clictx.DefaultContext())

	if err := c.Execute(); err != nil {
		os.Exit(1)
	}
}
