package addhdlrs

import (
	"github.com/spf13/pflag"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/cmds/ocm/common/options"
)

type ResourceSpecHandlerBase struct {
	options options.OptionSet
}

var (
	_ options.Options           = (*ResourceSpecHandlerBase)(nil)
	_ options.OptionSetProvider = (*ResourceSpecHandlerBase)(nil)
)

func NewBase(opts ...options.Options) ResourceSpecHandlerBase {
	return ResourceSpecHandlerBase{options: opts}
}

func (*ResourceSpecHandlerBase) RequireInputs() bool {
	return false
}

func (h *ResourceSpecHandlerBase) WithCLIOptions(opts ...options.Options) ResourceSpecHandlerBase {
	h.options = append(h.options, opts...)
	return *h
}

func (h *ResourceSpecHandlerBase) AsOptionSet() options.OptionSet {
	return h.options
}

func (h *ResourceSpecHandlerBase) AddFlags(opts *pflag.FlagSet) {
	h.options.AddFlags(opts)
}

func (h *ResourceSpecHandlerBase) GetTargetOpts() []ocm.TargetElementOption {
	return options.FindOptions[ocm.TargetElementOption](h.options)
}

func (h *ResourceSpecHandlerBase) GetElementModificationOpts() []ocm.ElementModificationOption {
	return options.FindOptions[ocm.ElementModificationOption](h.options)
}
