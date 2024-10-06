package plugin

import (
	"github.com/mandelsoft/goutils/generics"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm/extensions/attrs/plugincacheattr"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
)

func RegisterPlugins(ctx clictx.Context) {
	scheme := inputs.For(ctx)

	plugins := plugincacheattr.Get(ctx)
	for _, n := range plugins.PluginNames() {
		p := plugins.Get(n)
		if !p.IsValid() {
			continue
		}
		for _, d := range p.GetDescriptor().Inputs {
			t := NewType(d.Name, p, generics.Pointer(d))
			scheme.Register(t)
		}
	}
}
