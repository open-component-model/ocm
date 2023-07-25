// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/open-component-model/ocm/v2/cmds/helminstaller/app"
	"github.com/open-component-model/ocm/v2/pkg/contexts/clictx"
)

func main() {
	c := app.NewCliCommand(clictx.New(), nil)
	if err := c.Execute(); err != nil {
		os.Exit(1)
	}
}
