package plugin

import (
	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/options"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/labels/routingslip/spi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/descriptor"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type entryType struct {
	spi.EntryType
	plug    plugin.Plugin
	cliopts flagsets.ConfigOptionTypeSet
}

var _ spi.EntryType = (*entryType)(nil)

func NewType(name string, p plugin.Plugin, desc *plugin.ValueSetDescriptor) spi.EntryType {
	format := desc.Format
	if format != "" {
		format = "\n" + format
	}

	t := &entryType{
		plug: p,
	}

	cfghdlr := flagsets.NewConfigOptionTypeSetHandler(name, t.AddConfig)
	for _, o := range desc.CLIOptions {
		var opt flagsets.ConfigOptionType
		if o.Type == "" {
			opt = options.DefaultRegistry.GetOptionType(o.Name)
			if opt == nil {
				p.Context().Logger(plugin.TAG).Warn("unknown option", "plugin", p.Name(), "valueset", name, "option", o.Name)
			}
		} else {
			var err error
			opt, err = options.DefaultRegistry.CreateOptionType(o.Type, o.Name, o.Description)
			if err != nil {
				p.Context().Logger(plugin.TAG).Warn("invalid option", "plugin", p.Name(), "valueset", name, "option", o.Name, "error", err.Error())
			}
		}
		if opt != nil {
			cfghdlr.AddOptionType(opt)
		}
	}
	aopts := []spi.EntryTypeOption{spi.WithDescription(desc.Description), spi.WithFormatSpec(format)}
	if cfghdlr.Size() > 0 {
		aopts = append(aopts, spi.WithConfigHandler(cfghdlr))
		t.cliopts = cfghdlr
	}
	t.EntryType = spi.NewEntryType[*Entry](name, aopts...)
	return t
}

func (t *entryType) Decode(data []byte, unmarshaler runtime.Unmarshaler) (spi.Entry, error) {
	spec, err := t.EntryType.Decode(data, unmarshaler)
	if err != nil {
		return nil, err
	}
	spec.(*Entry).handler = NewPluginHandler(t.plug)
	return spec, nil
}

func (t *entryType) AddConfig(opts flagsets.ConfigOptions, cfg flagsets.Config) error {
	opts = opts.FilterBy(t.cliopts.HasOptionType)
	return t.plug.ComposeValueSet(descriptor.PURPOSE_ROUTINGSLIP, t.GetType(), opts, cfg)
}
