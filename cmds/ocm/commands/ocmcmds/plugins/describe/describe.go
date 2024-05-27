package describe

import (
	"encoding/json"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin"
	plugincommon "github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/common"
)

func DescribePlugin(p plugin.Plugin, out common.Printer) {
	out.Printf("Plugin Name:      %s\n", p.Name())
	out.Printf("Plugin Version:   %s\n", p.Version())
	out.Printf("BasePath:             %s\n", p.Path())

	if !p.IsValid() {
		out.Printf("Status:           %s\n", p.Error())
		return
	}
	out.Printf("Status:           %s\n", "valid")
	d := p.GetDescriptor()
	src := p.GetSource()
	if src != nil {
		out.Printf("Source:\n")
		out.Printf("  Component:       %s\n", src.Component)
		out.Printf("  Version:         %s\n", src.Version)
		out.Printf("  Resource:        %s\n", src.Resource)
		u := src.Repository.AsUniformSpec(p.Context())
		data, err := json.Marshal(src.Repository)
		if err != nil {
			out.Printf("Status:           %s\n", err)

			return
		}
		out.Printf("  Repository:      %s\n", u.String())
		out.Printf("    Specification: %s\n", string(data))
	} else {
		out.Printf("Source:           manually installed\n")
	}
	plugincommon.DescribePluginDescriptorCapabilities(p.Context().GetActions().GetActionTypes(), d, out)
}
