package plugin

import (
	"ocm.software/ocm/api/ocm/plugin"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/options"
)

type inputType struct {
	inputs.InputType
	plug    plugin.Plugin
	cliopts flagsets.ConfigOptionTypeSet
}

var _ inputs.InputType = (*inputType)(nil)

func NewType(name string, p plugin.Plugin, desc *plugin.InputTypeDescriptor) inputs.InputType {
	t := &inputType{
		plug: p,
	}

	cfghdlr := flagsets.NewConfigOptionTypeSetHandler(name, t.AddConfig)
	for _, o := range desc.CLIOptions {
		var opt flagsets.ConfigOptionType
		if o.Type == "" {
			opt = options.DefaultRegistry.GetOptionType(o.Name)
			if opt == nil {
				p.Context().Logger(plugin.TAG).Warn("unknown option", "plugin", p.Name(), "inputtype", name, "option", o.Name)
			}
		} else {
			var err error
			opt, err = options.DefaultRegistry.CreateOptionType(o.Type, o.Name, o.Description)
			if err != nil {
				p.Context().Logger(plugin.TAG).Warn("invalid option", "plugin", p.Name(), "inputtype", name, "option", o.Name, "error", err.Error())
			}
		}
		if opt != nil {
			cfghdlr.AddOptionType(opt)
		}
	}
	if cfghdlr.Size() > 0 {
		t.cliopts = cfghdlr
	}

	usage := desc.Description
	format := desc.Format
	if format != "" {
		usage += "\n" + format
	}
	t.InputType = inputs.NewInputType(name, &Spec{}, usage, cfghdlr)
	return t
}

func (t *inputType) Decode(data []byte, unmarshaler runtime.Unmarshaler) (inputs.InputSpec, error) {
	spec, err := t.InputType.Decode(data, unmarshaler)
	if err != nil {
		return nil, err
	}
	spec.(*Spec).handler = NewPluginHandler(t.plug)
	return spec, nil
}

func (t *inputType) AddConfig(opts flagsets.ConfigOptions, cfg flagsets.Config) error {
	opts = opts.FilterBy(t.cliopts.HasOptionType)
	return t.plug.ComposeInputSpec(t.GetType(), opts, cfg)
}
