// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/action/api"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/options"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/descriptor"
	"github.com/open-component-model/ocm/pkg/runtime"
	utils2 "github.com/open-component-model/ocm/pkg/utils"
)

func DescribePluginDescriptor(reg api.ActionTypeRegistry, d *descriptor.Descriptor, out common.Printer) {
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
		DescribeAccessMethods(d, out)
	}
	if len(d.Uploaders) > 0 {
		out.Printf("\n")
		// a working type inference would be really great
		ListElements[descriptor.UploaderDescriptor, descriptor.UploaderKey]("Repository Uploaders", d.Uploaders, out)
	}
	if len(d.Downloaders) > 0 {
		out.Printf("\n")
		ListElements[descriptor.DownloaderDescriptor, descriptor.DownloaderKey]("Resource Downloaders", d.Downloaders, out)
	}
	if len(d.Actions) > 0 {
		out.Printf("\n")
		out.Printf("Actions:\n")
		DescribeActions(reg, d, out)
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

func GetAccessMethodInfo(methods []descriptor.AccessMethodDescriptor) map[string]*MethodInfo {
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

func DescribeAccessMethods(d *descriptor.Descriptor, out common.Printer) {
	methods := GetAccessMethodInfo(d.AccessMethods)

	for _, n := range utils2.StringMapKeys(methods) {
		out.Printf("- Name: %s\n", n)
		m := methods[n]
		if m.Description != "" {
			out.Printf("%s\n", utils2.IndentLines(m.Description, "    "))
		}
		out := out.AddGap("  ") //nolint: govet // just use always out
		out.Printf("Versions:\n")
		for _, vn := range utils2.StringMapKeys(m.Versions) {
			out.Printf("- Version: %s\n", vn)
			out := out.AddGap("  ") //nolint: govet // just use always out
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

type ActionInfo struct {
	ActionDesc    string
	Versions      []string
	Selectors     []string
	ConsumerType  string
	Attributes    []string
	Description   string
	Usage         string
	KnownVersions []string
	BestVersion   string
	Error         string
}

func GetActionInfo(reg api.ActionTypeRegistry, actions []descriptor.ActionDescriptor) map[string]*ActionInfo {
	found := map[string]*ActionInfo{}
	for _, a := range actions {
		i := found[a.Name]
		if i == nil {
			i = &ActionInfo{
				ActionDesc:   a.Description,
				Versions:     append(a.Versions[:0:0], a.Versions...),
				Selectors:    append(a.DefaultSelectors[:0:0], a.DefaultSelectors...),
				ConsumerType: a.ConsumerType,
			}
			if err := runtime.SortVersions(i.Versions); err != nil {
				sort.Strings(i.Versions)
			}
			sort.Strings(i.Selectors)
			found[a.Name] = i
		}
		ad := reg.GetAction(a.Name)
		if ad == nil {
			i.Error = " (action unknown)"
		} else {
			i.Description = ad.Description()
			i.Usage = ad.Usage()
			i.KnownVersions = reg.SupportedActionVersions(a.Name)
			i.Attributes = ad.ConsumerAttributes()
			for _, v := range i.KnownVersions {
				for _, f := range a.Versions {
					if v == f {
						i.BestVersion = v
						break
					}
				}
			}
		}
	}
	return found
}

func DescribeActions(reg api.ActionTypeRegistry, d *descriptor.Descriptor, out common.Printer) {
	actions := GetActionInfo(reg, d.Actions)

	for _, n := range utils2.StringMapKeys(actions) {
		a := actions[n]
		out.Printf("- Name: %s%s\n", n, a.Error)
		if a.Description != "" {
			out.Printf("%s\n", utils2.IndentLines(a.Description, "    "))
		}
		if a.Usage != "" {
			out.Printf("\n%s\n", utils2.IndentLines(a.Usage, "    "))
		}
		if a.ActionDesc != "" {
			out.Printf("  Info:\n")
			out.Printf("%s\n", utils2.IndentLines(a.ActionDesc, "    "))
		}
		out := out.AddGap("  ") //nolint: govet // just use always out
		if a.BestVersion == "" {
			out.Printf("No version matches actual ocm version!\n")
		}
		out.Printf("Versions:\n")
		for _, vn := range a.Versions {
			_, err := semver.NewVersion(vn)
			switch {
			case err != nil:
				out.Printf("- %s (%s)\n", vn, err.Error())
			case vn == a.BestVersion:
				out.Printf("- %s (best matching)\n", vn)
			default:
				msg := " (not supported)"
				for _, v := range a.KnownVersions {
					if v == vn {
						msg = ""
					}
				}
				out.Printf("- %s%s\n", vn, msg)
			}
		}
		if a.ConsumerType == "" {
			out.Printf("Handler accepts standard credentials\n")
		} else {
			out.Printf("Consumer type: %s (consumer attributes described by action type)\n", a.ConsumerType)
			for _, p := range a.Attributes {
				out.Printf("- %s\n", p)
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

// ListElements lists describable elements.
func ListElements[E DescribableElement[C], C Describable](msg string, elems []E, out common.Printer) {
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
		out := out.AddGap("  ") //nolint: govet // just use always out
		if m.GetDescription() != "" {
			desc := m.GetDescription()
			if !strings.HasSuffix(desc, "\n") {
				desc += "\n"
			}
			out.Printf("%s\n", desc)
		}
		if len(m.GetConstraints()) > 0 {
			out.Printf("Registration Contraints:\n")
			for _, c := range m.GetConstraints() {
				out.Printf("- %s\n", utils2.IndentLines(c.Describe(), "  ", true))
			}
		}
	}
}
