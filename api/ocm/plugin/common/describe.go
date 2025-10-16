package common

import (
	"slices"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/mandelsoft/goutils/set"
	"ocm.software/ocm/api/datacontext/action/api"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/options"
	"ocm.software/ocm/api/ocm/plugin/descriptor"
	"ocm.software/ocm/api/utils"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/semverutils"
)

func DescribePluginDescriptor(reg api.ActionTypeRegistry, d *descriptor.Descriptor, out common.Printer) {
	out.Printf("Plugin Name:      %s\n", d.PluginName)
	out.Printf("Plugin Version:   %s\n", d.PluginVersion)
	DescribePluginDescriptorCapabilities(reg, d, out)
}

func DescribePluginDescriptorCapabilities(reg api.ActionTypeRegistry, d *descriptor.Descriptor, out common.Printer) {
	caps := d.Capabilities()
	if len(caps) == 0 {
		out.Printf("Capabilities:     none\n")
	} else {
		out.Printf("Capabilities:     %s\n", strings.Join(caps, ", "))
	}
	out.Printf("Description: \n")
	if d.Long == "" {
		out.Printf("%s\n", utils.IndentLines(d.Short, "      "))
	} else {
		out.Printf("%s\n", utils.IndentLines(d.Long, "      "))
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
	if len(d.ValueSets) > 0 {
		out.Printf("\n")
		out.Printf("Value Sets:\n")
		DescribeValueSets(d, out)
	}

	if len(d.ValueMergeHandlers) > 0 {
		out.Printf("\n")
		out.Printf("Value Merge Handlers:\n")
		DescribeValueMergeHandlers(d, out)
	}
	if len(d.ValueMergeHandlers) > 0 {
		out.Printf("\n")
		out.Printf("Label Merge Specifications:\n")
		DescribeLabelMergeSpecifications(d, out)
	}
	if len(d.Commands) > 0 {
		out.Printf("\n")
		out.Printf("CLI Extensions:\n")
		DescribeCLIExtensions(d, out)
	}
	if len(d.ConfigTypes) > 0 {
		out.Printf("\n")
		out.Printf("Config Types for CLI Command Extensions:\n")
		DescribeConfigTypes(d, out)
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

	for _, n := range utils.StringMapKeys(methods) {
		out.Printf("- Name: %s\n", n)
		m := methods[n]
		if m.Description != "" {
			out.Printf("%s\n", utils.IndentLines(m.Description, "    "))
		}
		out := out.AddGap("  ")
		out.Printf("Versions:\n")
		for _, vn := range utils.StringMapKeys(m.Versions) {
			out.Printf("- Version: %s\n", vn)
			out := out.AddGap("  ")
			v := m.Versions[vn]
			if v.Format != "" {
				out.Printf("%s\n", v.Format)
			}
			if len(v.Options) > 0 {
				out.Printf("Command Line Options:")
				out.Printf("%s\n", utils.FormatMap("", v.Options))
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
				Versions:     slices.Clone(a.Versions),
				Selectors:    slices.Clone(a.DefaultSelectors),
				ConsumerType: a.ConsumerType,
			}
			if err := semverutils.SortVersions(i.Versions); err != nil {
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

	for _, n := range utils.StringMapKeys(actions) {
		a := actions[n]
		out.Printf("- Name: %s%s\n", n, a.Error)
		if a.Description != "" {
			out.Printf("%s\n", utils.IndentLines(a.Description, "    "))
		}
		if a.Usage != "" {
			out.Printf("\n%s\n", utils.IndentLines(a.Usage, "    "))
		}
		if a.ActionDesc != "" {
			out.Printf("  Info:\n")
			out.Printf("%s\n", utils.IndentLines(a.ActionDesc, "    "))
		}
		out := out.AddGap("  ")
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

func DescribeValueMergeHandlers(d *descriptor.Descriptor, out common.Printer) {
	handlers := map[string]descriptor.ValueMergeHandlerDescriptor{}
	for _, h := range d.ValueMergeHandlers {
		handlers[h.GetName()] = h
	}

	for _, n := range utils.StringMapKeys(handlers) {
		a := handlers[n]
		out.Printf("- Name: %s\n", n)
		if a.Description != "" {
			out.Printf("%s\n", utils.IndentLines(a.Description, "    "))
		}
	}
}

func DescribeLabelMergeSpecifications(d *descriptor.Descriptor, out common.Printer) {
	handlers := map[string]descriptor.LabelMergeSpecification{}
	for _, h := range d.LabelMergeSpecifications {
		handlers[h.GetName()] = h
	}

	for _, n := range utils.StringMapKeys(handlers) {
		a := handlers[n]
		out.Printf("- Name: %s\n", n)
		if a.Description != "" {
			out.Printf("  Algorithm: %s\n", a.Algorithm)
			if len(a.Config) > 0 {
				out.Printf("  Config: %s\n", string(a.Config))
			}
			if a.Description != "" {
				out.Printf("%s\n", utils.IndentLines(a.Description, "    "))
			}
		}
	}
}

type ValueSetInfo struct {
	Name        string
	Description string
	Purposes    set.Set[string]
	Versions    map[string]*ValueSetVersion
}

type ValueSetVersion struct {
	Name    string
	Format  string
	Options map[string]options.OptionType
}

func GetValueSetInfo(valuesets []descriptor.ValueSetDescriptor) map[string]*ValueSetInfo {
	found := map[string]*ValueSetInfo{}
	for _, m := range valuesets {
		i := found[m.Name]
		if i == nil {
			i = &ValueSetInfo{
				Name:        m.Name,
				Description: m.Description,
				Versions:    map[string]*ValueSetVersion{},
				Purposes:    set.New(m.Purposes...),
			}
			found[m.Name] = i
		} else {
			i.Purposes.Add(m.Purposes...)
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
			v = &ValueSetVersion{
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

func DescribeValueSets(d *descriptor.Descriptor, out common.Printer) {
	valuesets := GetValueSetInfo(d.ValueSets)

	for _, n := range utils.StringMapKeys(valuesets) {
		out.Printf("- Name: %s\n", n)
		m := valuesets[n]
		out.Printf("  Purposes: %s\n", strings.Join(m.Purposes.AsArray(), ", "))
		if m.Description != "" {
			out.Printf("%s\n", utils.IndentLines(m.Description, "    "))
		}
		out := out.AddGap("  ")
		out.Printf("Versions:\n")
		for _, vn := range utils.StringMapKeys(m.Versions) {
			out.Printf("- Version: %s\n", vn)
			out := out.AddGap("  ")
			v := m.Versions[vn]
			if v.Format != "" {
				out.Printf("%s\n", v.Format)
			}
			if len(v.Options) > 0 {
				out.Printf("Command Line Options:")
				out.Printf("%s\n", utils.FormatMap("", v.Options))
			}
		}
	}
}

func DescribeCLIExtensions(d *descriptor.Descriptor, out common.Printer) {
	handlers := map[string]descriptor.CommandDescriptor{}
	for _, h := range d.Commands {
		handlers[h.GetName()] = h
	}

	for _, n := range utils.StringMapKeys(handlers) {
		a := handlers[n]
		s := a.Short
		if s != "" {
			s = " (" + s + ")"
		}
		out.Printf("- Name:   %s%s\n", n, s)
		if a.Description != "" {
			if len(a.ObjectType) > 0 {
				out.Printf("  Object: %s\n", a.ObjectType)
			}
			if len(a.Verb) > 0 {
				out.Printf("  Verb:   %s\n", a.Verb)
			}
			if len(a.Realm) > 0 {
				out.Printf("  Realm:  %s\n", a.Realm)
			}
			if len(a.Usage) > 0 {
				usage := ""
				if a.Verb != "" {
					usage += " " + a.Verb
					if a.ObjectType != "" {
						usage += " " + a.ObjectType
					} else {
						usage += " " + a.Name
					}
				} else {
					usage += " " + a.Name
				}
				i := strings.Index(a.Usage, " ")
				if i > 0 {
					usage += a.Usage[i:]
				}
				out.Printf("  Usage:  %s\n", usage[1:])
			}
			if a.Description != "" {
				out.Printf("%s\n", utils.IndentLines(a.Description, "    "))
			}
			if a.Example != "" {
				out.Printf("  Example:\n")
				out.Printf("%s\n", utils.IndentLines(a.Example, "    "))
			}
		}
	}
}

type TypeInfo struct {
	Name        string
	Description string
	Versions    map[string]*TypeVersion
}

type TypeVersion struct {
	Name   string
	Format string
}

func GetTypeInfo(types []descriptor.ConfigTypeDescriptor) map[string]*TypeInfo {
	found := map[string]*TypeInfo{}
	for _, m := range types {
		i := found[m.Name]
		if i == nil {
			i = &TypeInfo{
				Name:        m.Name,
				Description: m.Description,
				Versions:    map[string]*TypeVersion{},
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
			v = &TypeVersion{
				Name: vers,
			}
			i.Versions[vers] = v
		}
		if v.Format == "" {
			v.Format = m.Format
		}
	}
	return found
}

func DescribeConfigTypes(d *descriptor.Descriptor, out common.Printer) {
	types := GetTypeInfo(d.ConfigTypes)

	for _, n := range utils.StringMapKeys(types) {
		out.Printf("- Name: %s\n", n)
		m := types[n]
		if m.Description != "" {
			out.Printf("%s\n", utils.IndentLines(m.Description, "    "))
		}
		out := out.AddGap("  ")
		out.Printf("Versions:\n")
		for _, vn := range utils.StringMapKeys(m.Versions) {
			out.Printf("- Version: %s\n", vn)
			out := out.AddGap("  ")
			v := m.Versions[vn]
			if v.Format != "" {
				out.Printf("%s\n", v.Format)
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
	for _, n := range utils.StringMapKeys(keys) {
		m := keys[n]
		out.Printf("- Name: %s\n", n)
		out := out.AddGap("  ")
		if m.GetDescription() != "" {
			desc := m.GetDescription()
			if !strings.HasSuffix(desc, "\n") {
				desc += "\n"
			}
			out.Printf("%s\n", desc)
		}
		if len(m.GetConstraints()) > 0 {
			out.Printf("Registration Constraints:\n")
			for _, c := range m.GetConstraints() {
				out.Printf("- %s\n", utils.IndentLines(c.Describe(), "  ", true))
			}
		}
	}
}
