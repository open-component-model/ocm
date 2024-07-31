package addhdlrs

import (
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
)

type ResourceSpecHandlerBase struct {
	options options.OptionSet
}

var _ options.Options = (*ResourceSpecHandlerBase)(nil)

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

func (h *ResourceSpecHandlerBase) GetOptions() options.OptionSet {
	return h.options
}

func (h *ResourceSpecHandlerBase) AddFlags(opts *pflag.FlagSet) {
	h.options.AddFlags(opts)
}

func (h *ResourceSpecHandlerBase) GetTargetOpts() []ocm.TargetOption {
	return options.FindOptions[ocm.TargetOption](h.options)
}
