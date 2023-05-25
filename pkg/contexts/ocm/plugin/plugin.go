// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/identity/hostpath"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/cache"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/config"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/accessmethod"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/accessmethod/compose"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/accessmethod/get"
	accval "github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/accessmethod/validate"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/action"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/action/execute"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/download"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/upload"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/upload/put"
	uplval "github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/upload/validate"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type Plugin = *pluginImpl

var _ config.Target = (*pluginImpl)(nil)

type impl = cache.Plugin

// //nolint: errname // is no error.
type pluginImpl struct {
	lock sync.RWMutex
	ctx  ocm.Context
	impl
	config json.RawMessage
}

func NewPlugin(ctx ocm.Context, impl cache.Plugin, config json.RawMessage) Plugin {
	return &pluginImpl{
		ctx:    ctx,
		impl:   impl,
		config: config,
	}
}

func (p *pluginImpl) Context() ocm.Context {
	return p.ctx
}

func (p *pluginImpl) ConfigurePlugin(name string, data json.RawMessage) {
	if name == p.Name() {
		p.config = data
	}
}

func (p *pluginImpl) SetConfig(config json.RawMessage) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.config = config
}

func (p *pluginImpl) Action(spec ppi.ActionSpec, creds json.RawMessage) (ppi.ActionResult, error) {
	desc := p.GetActionDescriptor(spec.GetKind())
	if desc == nil {
		return nil, errors.ErrNotSupported(KIND_ACTION, spec.GetKind(), KIND_PLUGIN, p.Name())
	}
	if desc.ConsumerType != "" {
		cid := spec.GetConsumerAttributes()
		cid[cpi.ID_TYPE] = desc.ConsumerType
		c, err := credentials.CredentialsForConsumer(p.Context(), credentials.ConsumerIdentity(cid), hostpath.Matcher)
		if err != nil || c == nil {
			return nil, errors.ErrNotFound(credentials.KIND_CREDENTIALS, cid.String())
		}
		creds, err = json.Marshal(c.Properties())
		if err != nil {
			return nil, errors.Wrapf(err, "cannot marshal credentials")
		}
	}

	data, err := p.ctx.GetActions().GetActionTypes().EncodeActionSpec(spec, runtime.DefaultJSONEncoding)
	if err != nil {
		return nil, err
	}

	args := []string{action.Name, execute.Name, string(data)}
	if creds != nil {
		args = append(args, "--"+get.OptCreds, string(creds))
	}

	result, err := p.Exec(nil, nil, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "plugin %s", p.Name())
	}

	info, err := p.ctx.GetActions().GetActionTypes().DecodeActionResult(result, runtime.DefaultJSONEncoding)
	if err != nil {
		return nil, errors.Wrapf(err, "plugin %s: cannot unmarshal action result", p.Name())
	}
	return info, nil
}

func (p *pluginImpl) ValidateAccessMethod(spec []byte) (*ppi.AccessSpecInfo, error) {
	result, err := p.Exec(nil, nil, accessmethod.Name, accval.Name, string(spec))
	if err != nil {
		return nil, errors.Wrapf(err, "plugin %s", p.Name())
	}

	var info ppi.AccessSpecInfo
	err = json.Unmarshal(result, &info)
	if err != nil {
		return nil, errors.Wrapf(err, "plugin %s: cannot unmarshal access spec info", p.Name())
	}
	return &info, nil
}

func (p *pluginImpl) ComposeAccessMethod(name string, opts flagsets.ConfigOptions, base flagsets.Config) error {
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
		return errors.Wrapf(err, "cannot marshal access specification base value")
	}
	result, err := p.Exec(nil, nil, accessmethod.Name, compose.Name, name, string(optsdata), string(basedata))
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

func (p *pluginImpl) ValidateUploadTarget(name string, spec []byte) (*ppi.UploadTargetSpecInfo, error) {
	result, err := p.Exec(nil, nil, upload.Name, uplval.Name, name, string(spec))
	if err != nil {
		return nil, errors.Wrapf(err, "plugin uploader %s/%s", p.Name(), name)
	}

	var info ppi.UploadTargetSpecInfo
	err = json.Unmarshal(result, &info)
	if err != nil {
		return nil, errors.Wrapf(err, "plugin uploader %s/%s: cannot unmarshal upload target info", p.Name(), name)
	}
	return &info, nil
}

func (p *pluginImpl) Get(w io.Writer, creds, spec json.RawMessage) error {
	args := []string{accessmethod.Name, get.Name, string(spec)}
	if creds != nil {
		args = append(args, "--"+get.OptCreds, string(creds))
	}
	_, err := p.Exec(nil, w, args...)
	return err
}

func (p *pluginImpl) Put(name string, r io.Reader, artType, mimeType, hint string, creds, target json.RawMessage) (ocm.AccessSpec, error) {
	args := []string{upload.Name, put.Name, name, string(target)}

	if creds != nil {
		args = append(args, "--"+put.OptCreds, string(creds))
	}
	if hint != "" {
		args = append(args, "--"+put.OptHint, hint)
	}
	if mimeType != "" {
		args = append(args, "--"+put.OptMedia, mimeType)
	}
	if artType != "" {
		args = append(args, "--"+put.OptArt, artType)
	}
	result, err := p.Exec(r, nil, args...)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	err = json.Unmarshal(result, &m)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal put result")
	}
	if len(m) == 0 {
		return nil, nil // not used
	}
	return p.ctx.AccessSpecForConfig(result, runtime.DefaultJSONEncoding)
}

func (p *pluginImpl) Download(name string, r io.Reader, artType, mimeType, target string) (bool, string, error) {
	args := []string{download.Name, name, target}

	if mimeType != "" {
		args = append(args, "--"+put.OptMedia, mimeType)
	}
	if artType != "" {
		args = append(args, "--"+put.OptArt, artType)
	}
	result, err := p.Exec(r, nil, args...)
	if err != nil {
		return true, "", err
	}
	var m download.Result
	err = json.Unmarshal(result, &m)
	if err != nil {
		return true, "", errors.Wrapf(err, "cannot unmarshal put result")
	}
	if m.Error != "" {
		return true, "", fmt.Errorf("%s", m.Error)
	}
	return m.Path != "", m.Path, nil
}

func (p *pluginImpl) Exec(r io.Reader, w io.Writer, args ...string) ([]byte, error) {
	if len(p.config) == 0 {
		p.ctx.Logger(TAG).Debug("execute plugin action", "path", p.Path(), "args", args)
	} else {
		p.ctx.Logger(TAG).Debug("execute plugin action", "path", p.Path(), "args", args, "config", p.config)
	}
	return cache.Exec(p.Path(), p.config, r, w, args...)
}
