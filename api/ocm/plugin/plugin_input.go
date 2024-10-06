package plugin

import (
	"encoding/json"
	"io"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/input"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/input/compose"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/input/get"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/input/validate"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
)

func (p *pluginImpl) ValidateInputSpec(spec []byte) (*ppi.InputSpecInfo, error) {
	result, err := p.Exec(nil, nil, input.Name, validate.Name, string(spec))
	if err != nil {
		return nil, errors.Wrapf(err, "plugin %s", p.Name())
	}

	var info ppi.InputSpecInfo
	err = json.Unmarshal(result, &info)
	if err != nil {
		return nil, errors.Wrapf(err, "plugin %s: cannot unmarshal input spec info", p.Name())
	}
	return &info, nil
}

func (p *pluginImpl) ComposeInputSpec(name string, opts flagsets.ConfigOptions, base flagsets.Config) error {
	cfg := flagsets.Config{}
	for _, o := range opts.Options() {
		cfg[o.GetName()] = o.Value()
	}
	optsdata, err := json.Marshal(cfg)
	if err != nil {
		return errors.Wrapf(err, "cannot marshal option values")
	}
	basedata, err := json.Marshal(base)
	if err != nil {
		return errors.Wrapf(err, "cannot marshal input specification base value")
	}
	result, err := p.Exec(nil, nil, input.Name, compose.Name, name, string(optsdata), string(basedata))
	if err != nil {
		return err
	}
	var r flagsets.Config
	err = json.Unmarshal(result, &r)
	if err != nil {
		return errors.Wrapf(err, "cannot unmarshal composition result")
	}

	for k := range base {
		delete(base, k)
	}
	for k, v := range r {
		base[k] = v
	}
	return nil
}

func (p *pluginImpl) GetInputBlob(w io.Writer, creds, spec json.RawMessage) error {
	args := []string{input.Name, get.Name, string(spec)}
	if creds != nil {
		args = append(args, "--"+get.OptCreds, string(creds))
	}
	_, err := p.Exec(nil, w, args...)
	return err
}
