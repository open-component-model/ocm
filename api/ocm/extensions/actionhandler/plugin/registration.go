package plugin

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/datacontext/action/handlers"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/attrs/plugincacheattr"
	"ocm.software/ocm/api/ocm/plugin"
	"ocm.software/ocm/api/utils/registrations"
)

func init() {
	handlers.DefaultRegistry().RegisterRegistrationHandler("plugin", &RegistrationHandler{})
}

type RegistrationHandler struct{}

var _ handlers.HandlerRegistrationHandler = (*RegistrationHandler)(nil)

func (r *RegistrationHandler) RegisterByName(handler string, target handlers.Target, config handlers.HandlerConfig, olist ...handlers.Option) (bool, error) {
	path := cpi.NewNamePath(handler)

	if config == nil {
		return true, fmt.Errorf("config required")
	}

	ctx, ok := config.(cpi.Context)
	if !ok {
		return true, fmt.Errorf("expected ocm.Context as config but found: %T", config)
	}
	if len(path) != 1 {
		return true, fmt.Errorf("plugin handler must be of the form <plugin>")
	}

	opts := handlers.NewOptions(olist...)
	name := path[0]
	err := RegisterActionHandler(target, name, ctx, opts)
	return true, err
}

func RegisterActionHandler(target handlers.Target, pname string, ctx ocm.Context, opts *handlers.Options) error {
	set := plugincacheattr.Get(ctx)
	if set == nil {
		return errors.ErrUnknown(plugin.KIND_PLUGIN, pname)
	}

	p := set.Get(pname)
	if p == nil {
		return errors.ErrUnknown(plugin.KIND_PLUGIN, pname)
	}

	h, err := New(p, opts.Action)
	if err != nil {
		return err
	}
	return target.GetActions().Register(h, opts)
}

func (r *RegistrationHandler) GetHandlers(target handlers.Target) registrations.HandlerInfos {
	infos := registrations.HandlerInfos{}

	ctx := ocm.DefaultContext()
	if c, ok := target.(ocm.ContextProvider); ok {
		ctx = c.OCMContext()
	}

	set := plugincacheattr.Get(ctx)
	if set == nil {
		return infos
	}

	for _, name := range set.PluginNames() {
		for _, a := range set.Get(name).GetDescriptor().Actions {
			d := target.GetActions().GetActionTypes().GetAction(a.GetName())
			short := ""
			if d != nil {
				short = d.Description()
			}
			i := registrations.HandlerInfo{
				Name:        name + "/" + a.GetName(),
				ShortDesc:   short,
				Description: a.GetDescription(),
			}
			infos = append(infos, i)
		}
	}
	return infos
}
