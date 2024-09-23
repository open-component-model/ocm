package plugin

import (
	"fmt"
	"slices"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/attrs/plugincacheattr"
	"ocm.software/ocm/api/ocm/plugin"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/standard"
	"ocm.software/ocm/api/utils/registrations"
)

type Config = interface{}

func init() {
	transferhandler.RegisterHandlerRegistrationHandler("plugin", &RegistrationHandler{})
}

type RegistrationHandler struct{}

var _ transferhandler.ByNameCreationHandler = (*RegistrationHandler)(nil)

func (r *RegistrationHandler) ByName(ctx ocm.Context, handler string, opts ...transferhandler.TransferOption) (bool, transferhandler.TransferHandler, error) {
	path := cpi.NewNamePath(handler)

	if len(path) < 1 || len(path) > 2 {
		return true, nil, fmt.Errorf("plugin handler name must be of the form <plugin>[/<downloader>]")
	}

	name := ""
	if len(path) > 1 {
		name = path[1]
	}

	h, err := CreateTransferHandler(ctx, path[0], name, opts...)
	return true, h, err
}

func CreateTransferHandler(ctx ocm.Context, pname, name string, opts ...transferhandler.TransferOption) (transferhandler.TransferHandler, error) {
	options := &Options{}
	err := transferhandler.ApplyOptions(options, append(slices.Clone(opts), TransferHandler(name), Plugin(pname))...)
	if err != nil {
		return nil, err
	}

	if options.plugin != "" && options.plugin != pname {
		return nil, fmt.Errorf("plugin option not possible for path-based transferhandler creation")
	}
	if options.handler != "" && options.handler != name {
		return nil, fmt.Errorf("plugin transfer handler option nor possible for path-based transferhandler creation")
	}

	set := plugincacheattr.Get(ctx)
	if set == nil {
		return nil, errors.ErrUnknown(plugin.KIND_PLUGIN, pname)
	}

	p := set.Get(pname)
	if p == nil {
		return nil, errors.ErrUnknown(plugin.KIND_PLUGIN, pname)
	}
	d := p.GetTransferHandler(name)
	if d == nil {
		return nil, errors.ErrNotFound(plugin.KIND_TRANSFERHANDLER, name, pname)
	}

	return &Handler{
		Handler: *standard.NewDefaultHandler(&options.Options),
		opts:    options,
		plugin:  p,
		desc:    d,
	}, nil
}

func (r *RegistrationHandler) GetHandlers(target *transferhandler.Target) registrations.HandlerInfos {
	infos := registrations.NewNodeHandlerInfo("transfer handlers provided by plugins",
		"sub namespace of the form <code>&lt;plugin name>/&lt;handler></code>")

	set := plugincacheattr.Get(target.Context)
	if set == nil {
		return infos
	}

	for _, name := range set.PluginNames() {
		p := set.Get(name)
		if !p.IsValid() {
			continue
		}
		for _, d := range set.Get(name).GetDescriptor().Downloaders {
			i := registrations.HandlerInfo{
				Name:        name + "/" + d.GetName(),
				ShortDesc:   "",
				Description: d.GetDescription(),
			}
			infos = append(infos, i)
		}
	}
	return infos
}
