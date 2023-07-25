// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/open-component-model/ocm/v2/cmds/ecrplugin/actions"
	"github.com/open-component-model/ocm/v2/cmds/ecrplugin/config"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/plugin/ppi"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/plugin/ppi/cmds"
	"github.com/open-component-model/ocm/v2/pkg/version"
)

func main() {
	p := ppi.NewPlugin("ecrplugin", version.Get().String())

	p.SetShort("AWS ecr repository creation")
	p.SetLong("plugin assuring the existence of required AWS ECR repositories")
	p.SetConfigParser(config.GetConfig)
	p.SetDescriptorTweaker(func(d ppi.Descriptor) ppi.Descriptor {
		cfg, _ := p.GetConfig()
		if cfg == nil {
			return d
		}
		return config.TweakDescriptor(d, cfg.(*config.Config))
	})

	p.RegisterAction(actions.New())
	err := cmds.NewPluginCommand(p).Execute(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}
}
