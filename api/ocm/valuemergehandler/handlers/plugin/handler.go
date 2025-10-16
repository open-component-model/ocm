package plugin

import (
	"encoding/json"

	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/ocm/plugin"
	"ocm.software/ocm/api/ocm/plugin/descriptor"
	"ocm.software/ocm/api/ocm/valuemergehandler/hpi"
)

// pluginHandler delegates action to a plugin based handler.
type pluginHandler struct {
	plugin     plugin.Plugin
	descriptor *descriptor.ValueMergeHandlerDescriptor
}

func New(p plugin.Plugin, name string) (hpi.Handler, error) {
	md := p.GetValueMergeHandlerDescriptor(name)
	if md == nil {
		return nil, errors.ErrUnknown(hpi.KIND_VALUE_MERGE_ALGORITHM, name, plugin.KIND_PLUGIN, p.Name())
	}

	return &pluginHandler{
		plugin:     p,
		descriptor: md,
	}, nil
}

func (b *pluginHandler) Algorithm() string {
	return b.descriptor.Name
}

func (b *pluginHandler) Description() string {
	return b.descriptor.Description
}

func (b *pluginHandler) DecodeConfig(data []byte) (hpi.Config, error) {
	var cfg Config
	err := json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (b *pluginHandler) Merge(_ hpi.Context, src hpi.Value, tgt *hpi.Value, cfg hpi.Config) (bool, error) {
	spec, err := hpi.NewSpecification(b.descriptor.Name, cfg)
	if err != nil {
		return false, err
	}
	mod, r, err := b.plugin.MergeValue(spec, src, *tgt)
	if err != nil {
		return false, err
	}
	if mod {
		tgt.RawMessage = r.RawMessage
	}
	return mod, nil
}
