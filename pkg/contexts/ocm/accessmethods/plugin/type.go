package plugin

import (
	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/options"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi/accspeccpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type accessType struct {
	accspeccpi.AccessType
	plug    plugin.Plugin
	cliopts flagsets.ConfigOptionTypeSet
}

var _ accspeccpi.AccessType = (*accessType)(nil)

func NewType(name string, p plugin.Plugin, desc *plugin.AccessMethodDescriptor) accspeccpi.AccessType {
	format := desc.Format
	if format != "" {
		format = "\n" + format
	}

	t := &accessType{
		plug: p,
	}

	cfghdlr := flagsets.NewConfigOptionTypeSetHandler(name, t.AddConfig)
	for _, o := range desc.CLIOptions {
		var opt flagsets.ConfigOptionType
		if o.Type == "" {
			opt = options.DefaultRegistry.GetOptionType(o.Name)
			if opt == nil {
				p.Context().Logger(plugin.TAG).Warn("unknown option", "plugin", p.Name(), "accessmethod", name, "option", o.Name)
			}
		} else {
			var err error
			opt, err = options.DefaultRegistry.CreateOptionType(o.Type, o.Name, o.Description)
			if err != nil {
				p.Context().Logger(plugin.TAG).Warn("invalid option", "plugin", p.Name(), "accessmethod", name, "option", o.Name, "error", err.Error())
			}
		}
		if opt != nil {
			cfghdlr.AddOptionType(opt)
		}
	}
	aopts := []accspeccpi.AccessSpecTypeOption{accspeccpi.WithDescription(desc.Description), accspeccpi.WithFormatSpec(format)}
	if cfghdlr.Size() > 0 {
		aopts = append(aopts, accspeccpi.WithConfigHandler(cfghdlr))
		t.cliopts = cfghdlr
	}
	t.AccessType = accspeccpi.NewAccessSpecType[*AccessSpec](name, aopts...)
	return t
}

func (t *accessType) Decode(data []byte, unmarshaler runtime.Unmarshaler) (accspeccpi.AccessSpec, error) {
	spec, err := t.AccessType.Decode(data, unmarshaler)
	if err != nil {
		return nil, err
	}
	spec.(*AccessSpec).handler = NewPluginHandler(t.plug)
	return spec, nil
}

func (t *accessType) AddConfig(opts flagsets.ConfigOptions, cfg flagsets.Config) error {
	opts = opts.FilterBy(t.cliopts.HasOptionType)
	return t.plug.ComposeAccessMethod(t.GetType(), opts, cfg)
}
