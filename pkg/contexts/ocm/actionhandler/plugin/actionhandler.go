package plugin

import (
	"encoding/json"

	"github.com/mandelsoft/goutils/errors"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/action"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/action/handlers"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin"
)

// pluginHandler delegates action to a plugin based handler.
type pluginHandler struct {
	plugin plugin.Plugin
	name   string
}

func New(p plugin.Plugin, name string) (handlers.ActionHandler, error) {
	ad := p.GetActionDescriptor(name)
	if ad == nil {
		return nil, errors.ErrUnknown(plugin.KIND_ACTION, name, plugin.KIND_PLUGIN, p.Name())
	}

	return &pluginHandler{
		plugin: p,
		name:   name,
	}, nil
}

func (b *pluginHandler) Handle(spec action.ActionSpec, creds common.Properties) (action.ActionResult, error) {
	var err error
	var creddata json.RawMessage

	if len(creds) != 0 {
		creddata, err = json.Marshal(creds)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot marshal credentials")
		}
	}

	return b.plugin.Action(spec, creddata)
}
