package plugin

import (
	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/standard"
)

func init() {
	transferhandler.RegisterHandler(100, &TransferOptionsCreator{})
}

type Options struct {
	standard.Options
	plugin  string
	handler string
	config  interface{}
}

var (
	_ transferhandler.TransferOption = (*Options)(nil)

	_ PluginNameOption            = (*Options)(nil)
	_ TransferHandlerOption       = (*Options)(nil)
	_ TransferHandlerConfigOption = (*Options)(nil)
)

type TransferOptionsCreator = transferhandler.SpecializedOptionsCreator[*Options, Options]

func (o *Options) NewOptions() transferhandler.TransferHandlerOptions {
	return &Options{}
}

func (o *Options) NewTransferHandler() (transferhandler.TransferHandler, error) {
	return New(o)
}

func (o *Options) ApplyTransferOption(target transferhandler.TransferOptions) error {
	if o.plugin != "" {
		if opts, ok := target.(PluginNameOption); ok {
			opts.SetPluginName(o.plugin)
		}
	}
	if o.plugin != "" {
		if opts, ok := target.(TransferHandlerOption); ok {
			opts.SetTransferHandler(o.handler)
		}
	}
	if o.config != nil {
		if opts, ok := target.(TransferHandlerConfigOption); ok {
			opts.SetTransferHandlerConfig(o.config)
		}
	}
	return o.Options.ApplyTransferOption(target)
}

func (o *Options) SetPluginName(name string) {
	o.plugin = name
}

func (o *Options) GetPluginName() string {
	return o.plugin
}

func (o *Options) SetTransferHandler(name string) {
	o.handler = name
}

func (o *Options) GetTransferHandler() string {
	return o.handler
}

func (o *Options) SetTransferHandlerConfig(cfg interface{}) {
	o.config = cfg
}

func (o *Options) GetTransferHandlerConfig() interface{} {
	return o.config
}

///////////////////////////////////////////////////////////////////////////////

type PluginNameOption interface {
	SetPluginName(name string)
	GetPluginName() string
}

type pluginNameOption struct {
	TransferOptionsCreator
	name string
}

func (o *pluginNameOption) ApplyTransferOption(to transferhandler.TransferOptions) error {
	if eff, ok := to.(PluginNameOption); ok {
		eff.SetPluginName(o.name)
		return nil
	} else {
		return errors.ErrNotSupported(transferhandler.KIND_TRANSFEROPTION, "pluginName")
	}
}

func Plugin(name string) transferhandler.TransferOption {
	return &pluginNameOption{
		name: name,
	}
}

///////////////////////////////////////////////////////////////////////////////

type TransferHandlerOption interface {
	SetTransferHandler(name string)
	GetTransferHandler() string
}

type transferHandlerOption struct {
	TransferOptionsCreator
	name string
}

func (o *transferHandlerOption) ApplyTransferOption(to transferhandler.TransferOptions) error {
	if eff, ok := to.(TransferHandlerOption); ok {
		eff.SetTransferHandler(o.name)
		return nil
	} else {
		return errors.ErrNotSupported(transferhandler.KIND_TRANSFEROPTION, "transferHandler")
	}
}

func TransferHandler(name string) transferhandler.TransferOption {
	return &transferHandlerOption{
		name: name,
	}
}

///////////////////////////////////////////////////////////////////////////////

type TransferHandlerConfigOption interface {
	SetTransferHandlerConfig(cf interface{})
	GetTransferHandlerConfig() interface{}
}

type transferHandlerConfigOption struct {
	TransferOptionsCreator
	config interface{}
}

func (o *transferHandlerConfigOption) ApplyTransferOption(to transferhandler.TransferOptions) error {
	if eff, ok := to.(TransferHandlerConfigOption); ok {
		eff.SetTransferHandlerConfig(o.config)
		return nil
	} else {
		return errors.ErrNotSupported(transferhandler.KIND_TRANSFEROPTION, "transferHandlerConfig")
	}
}

func TransferHandlerConfig(cfg interface{}) transferhandler.TransferOption {
	return &transferHandlerConfigOption{
		config: cfg,
	}
}
