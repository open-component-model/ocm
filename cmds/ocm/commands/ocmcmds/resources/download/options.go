package download

import (
	"github.com/spf13/pflag"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/cmds/ocm/common/output"
)

func From(o *output.Options) *Option {
	var opt *Option
	o.Get(&opt)
	return opt
}

func NewOptions(silent ...bool) *Option {
	return &Option{SilentOption: utils.Optional(silent...)}
}

type Option struct {
	SilentOption bool
	UseHandlers  bool
	Verify       bool
}

func (o *Option) SetUseHandlers(ok ...bool) *Option {
	o.UseHandlers = utils.OptionalDefaultedBool(true, ok...)
	return o
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	if !o.SilentOption {
		fs.BoolVarP(&o.UseHandlers, "download-handlers", "d", false, "use download handler if possible")
	}
	fs.BoolVarP(&o.Verify, "verify", "", false, "verify downloads")
}

func (o *Option) Usage() string {
	s := `
The library supports some downloads with semantics based on resource types. For example a helm chart
can be download directly as helm chart archive, even if stored as OCI artifact.
This is handled by download handler. Their usage can be enabled with the <code>--download-handlers</code>
option. Otherwise the resource as returned by the access method is stored.
`
	return s
}
