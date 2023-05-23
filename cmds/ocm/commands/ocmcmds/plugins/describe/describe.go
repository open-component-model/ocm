// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package describe

import (
	"encoding/json"
	"strings"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/action/api"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin"
	plugincommon "github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/common"
	utils2 "github.com/open-component-model/ocm/pkg/utils"
)

func DescribePlugin(p plugin.Plugin, out common.Printer) {
	out.Printf("Plugin Name:      %s\n", p.Name())
	out.Printf("Plugin Version:   %s\n", p.Version())
	out.Printf("Path:             %s\n", p.Path())

	if !p.IsValid() {
		out.Printf("Status:           %s\n", p.Error())
		return
	}
	out.Printf("Status:           %s\n", "valid")
	var caps []string
	d := p.GetDescriptor()
	if len(d.AccessMethods) > 0 {
		caps = append(caps, "Access Methods")
	}
	if len(d.Uploaders) > 0 {
		caps = append(caps, "Repository Uploaders")
	}
	if len(d.Downloaders) > 0 {
		caps = append(caps, "Resource Downloaders")
	}
	if len(d.Actions) > 0 {
		caps = append(caps, "Actions")
	}
	if len(caps) == 0 {
		out.Printf("Capabilities:     none\n")
	} else {
		out.Printf("Capabilities:     %s\n", strings.Join(caps, ", "))
	}
	src := p.GetSource()
	if src != nil {
		out.Printf("Source:\n")
		out.Printf("  Component:       %s\n", src.Component)
		out.Printf("  Version:         %s\n", src.Version)
		out.Printf("  Resource:        %s\n", src.Resource)
		u := src.Repository.AsUniformSpec(p.Context())
		data, _ := json.Marshal(src.Repository)
		out.Printf("  Repository:      %s\n", u.String())
		out.Printf("    Specification: %s\n", string(data))
	} else {
		out.Printf("Source:           manually installed\n")
	}
	out.Printf("Description: \n")
	if d.Long == "" {
		out.Printf("%s\n", utils2.IndentLines(d.Short, "      "))
	} else {
		out.Printf("%s\n", utils2.IndentLines(d.Long, "      "))
	}
	if len(d.AccessMethods) > 0 {
		out.Printf("Access Methods:\n")
		plugincommon.DescribeAccessMethods(d, out)
	}
	if len(d.Uploaders) > 0 {
		// a working type inference would be really great
		plugincommon.ListElements[plugin.UploaderDescriptor, plugin.UploaderKey]("Repository Uploaders", d.Uploaders, out)
	}
	if len(d.Downloaders) > 0 {
		plugincommon.ListElements[plugin.DownloaderDescriptor, plugin.DownloaderKey]("Resource Downloaders", d.Downloaders, out)
	}
	if len(d.Actions) > 0 {
		out.Printf("Actions:\n")
		plugincommon.DescribeActions(p.Context().GetActions().GetActionTypes(), d, out)
	}
}

func DescribePluginDescriptor(reg api.ActionTypeRegistry, d *plugin.Descriptor, out common.Printer) {
	out.Printf("Plugin Name:      %s\n", d.PluginName)
	out.Printf("Plugin Version:   %s\n", d.PluginVersion)

	var caps []string
	if len(d.AccessMethods) > 0 {
		caps = append(caps, "Access Methods")
	}
	if len(d.Uploaders) > 0 {
		caps = append(caps, "Repository Uploaders")
	}
	if len(d.Downloaders) > 0 {
		caps = append(caps, "Resource Downloaders")
	}
	if len(d.Actions) > 0 {
		caps = append(caps, "Actions")
	}
	if len(caps) == 0 {
		out.Printf("Capabilities:     none\n")
	} else {
		out.Printf("Capabilities:     %s\n", strings.Join(caps, ", "))
	}
	out.Printf("Description: \n")
	if d.Long == "" {
		out.Printf("%s\n", utils2.IndentLines(d.Short, "      "))
	} else {
		out.Printf("%s\n", utils2.IndentLines(d.Long, "      "))
	}
	if len(d.AccessMethods) > 0 {
		out.Printf("\n")
		out.Printf("Access Methods:\n")
		plugincommon.DescribeAccessMethods(d, out)
	}
	if len(d.Uploaders) > 0 {
		out.Printf("\n")
		// a working type inference would be really great
		plugincommon.ListElements[plugin.UploaderDescriptor, plugin.UploaderKey]("Repository Uploaders", d.Uploaders, out)
	}
	if len(d.Downloaders) > 0 {
		out.Printf("\n")
		plugincommon.ListElements[plugin.DownloaderDescriptor, plugin.DownloaderKey]("Resource Downloaders", d.Downloaders, out)
	}
	if len(d.Actions) > 0 {
		out.Printf("\n")
		out.Printf("Actions:\n")
		plugincommon.DescribeActions(reg, d, out)
	}
}
