// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/open-component-model/ocm/cmds/demoplugin/accessmethods"
	"github.com/open-component-model/ocm/cmds/demoplugin/config"
	"github.com/open-component-model/ocm/cmds/demoplugin/uploaders"
	"github.com/open-component-model/ocm/cmds/demoplugin/valuesets"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds"
	"github.com/open-component-model/ocm/pkg/version"
)

func main() {
	p := ppi.NewPlugin("demo", version.Get().String())

	p.SetShort("demo plugin")
	p.SetLong("plugin providing access to temp files and a check routing slip entry.")
	p.SetConfigParser(config.GetConfig)

	p.RegisterAccessMethod(accessmethods.New())
	u := uploaders.New()
	p.RegisterUploader("testArtifact", "", u)
	p.RegisterValueSet(valuesets.New())
	err := cmds.NewPluginCommand(p).Execute(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}
}
