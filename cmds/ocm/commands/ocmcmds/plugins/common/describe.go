// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"encoding/json"
	"strings"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/options"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin"
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
	if len(p.GetDescriptor().AccessMethods) > 0 {
		caps = append(caps, "Access Methods")
	}
	if len(p.GetDescriptor().Uploaders) > 0 {
		caps = append(caps, "Repository Uploaders")
	}
	if len(p.GetDescriptor().Downloaders) > 0 {
		caps = append(caps, "Resource Downloaders")
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
	out.Printf("\n")
	out.Printf("Description: \n")
	if p.GetDescriptor().Long == "" {
		out.Printf("%s\n", utils2.IndentLines(p.GetDescriptor().Short, "      "))
	} else {
		out.Printf("%s\n", utils2.IndentLines(p.GetDescriptor().Long, "      "))
	}
	if len(p.GetDescriptor().AccessMethods) > 0 {
		out.Printf("\n")
		out.Printf("Access Methods:\n")
		DescribeAccessMethods(p, out)
	}
	if len(p.GetDescriptor().Uploaders) > 0 {
		out.Printf("\n")
		// a working type inference would be really great
		ListElements[plugin.UploaderDescriptor, plugin.UploaderKey]("Repository Uploaders", p.GetDescriptor().Uploaders, out)
	}
	if len(p.GetDescriptor().Downloaders) > 0 {
		out.Printf("\n")
		ListElements[plugin.DownloaderDescriptor, plugin.DownloaderKey]("Resource Downloaders", p.GetDescriptor().Downloaders, out)
	}
}

type MethodInfo struct {
	Name        string
	Description string
	Versions    map[string]*MethodVersion
}

type MethodVersion struct {
	Name    string
	Format  string
	Options map[string]options.OptionType
}

func GetAccessMethodInfo(methods []plugin.AccessMethodDescriptor) map[string]*MethodInfo {
	found := map[string]*MethodInfo{}
	for _, m := range methods {
		i := found[m.Name]
		if i == nil {
			i = &MethodInfo{
				Name:        m.Name,
				Description: m.Description,
				Versions:    map[string]*MethodVersion{},
			}
			found[m.Name] = i
		}
		if i.Description == "" {
			i.Description = m.Description
		}
		vers := m.Version
		if m.Version == "" {
			vers = "v1"
		}
		v := i.Versions[vers]
		if v == nil {
			v = &MethodVersion{
				Name:    vers,
				Options: map[string]options.OptionType{},
			}
			i.Versions[vers] = v
		}
		if v.Format == "" {
			v.Format = m.Format
		}
		if (len(v.Options) == 0 || m.Version != "") && len(m.CLIOptions) > 0 {
			for _, o := range m.CLIOptions {
				if o.Name == "" {
					continue
				}
				opt := options.DefaultRegistry.GetOptionType(o.Name)
				if opt == nil {
					t, err := options.DefaultRegistry.CreateOptionType(o.Type, o.Name, o.Description)
					if err != nil {
						continue
					}
					opt = t
				}
				v.Options[opt.GetName()] = opt
			}
		}
	}
	return found
}

func DescribeAccessMethods(p plugin.Plugin, out common.Printer) {
	d := p.GetDescriptor()

	methods := GetAccessMethodInfo(d.AccessMethods)

	for _, n := range utils2.StringMapKeys(methods) {
		out.Printf("- Name: %s\n", n)
		m := methods[n]
		if m.Description != "" {
			out.Printf("%s\n", utils2.IndentLines(m.Description, "    "))
		}
		out := out.AddGap("  ")
		out.Printf("Versions:\n")
		for _, vn := range utils2.StringMapKeys(m.Versions) {
			out.Printf("- Version: %s\n", vn)
			out := out.AddGap("  ")
			v := m.Versions[vn]
			if v.Format != "" {
				out.Printf("%s\n", v.Format)
			}
			if len(v.Options) > 0 {
				out.Printf("Command Line Options:")
				out.Printf("%s\n", utils2.FormatMap("", v.Options))
			}
		}
	}
}

type Describable interface {
	Describe() string
}

type DescribableElement[C Describable] interface {
	GetName() string
	GetDescription() string
	GetConstraints() []C
}

func ListElements[E DescribableElement[C], C Describable](msg string, elems []E, out common.Printer) {
	var list []string

	keys := map[string]E{}
	for _, e := range elems {
		keys[e.GetName()] = e
	}
	if len(keys) > 0 {
		out.Printf("%s:\n", msg)
	}
	for _, n := range utils2.StringMapKeys(keys) {
		m := keys[n]
		out.Printf("- Name: %s\n", n)
		if m.GetDescription() != "" {
			desc := m.GetDescription()
			if !strings.HasSuffix(desc, "\n") {
				desc += "\n"
			}
			out.AddGap("  ").Printf("%s\n", desc)
		}
		if len(m.GetConstraints()) > 0 {
			out := out.AddGap("  ")
			out.Printf("Registration Contraints:\n")
			for _, c := range m.GetConstraints() {
				out.Printf("- %s\n", utils2.IndentLines(c.Describe(), "  ", true))
			}
		}
		list = append(list, n)
	}
}
