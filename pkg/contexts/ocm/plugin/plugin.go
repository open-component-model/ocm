// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/internal"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds"
	"github.com/open-component-model/ocm/pkg/errors"
)

type Plugin = *pluginImpl

// //nolint: errname // is no error.
type pluginImpl struct {
	name       string
	config     json.RawMessage
	descriptor *internal.Descriptor
	path       string
	error      string
}

func NewPlugin(name string, path string, config json.RawMessage, desc *Descriptor, errmsg string) Plugin {
	return &pluginImpl{
		name:       name,
		path:       path,
		config:     config,
		descriptor: desc,
		error:      errmsg,
	}
}

func (p *pluginImpl) GetDescriptor() *internal.Descriptor {
	return p.descriptor
}

func (p *pluginImpl) Name() string {
	return p.name
}

func (p *pluginImpl) Path() string {
	return p.path
}

func (p *pluginImpl) Version() string {
	if !p.IsValid() {
		return "-"
	}
	return p.descriptor.PluginVersion
}

func (p *pluginImpl) IsValid() bool {
	return p.descriptor != nil
}

func (p *pluginImpl) Error() string {
	return p.error
}

func (p *pluginImpl) SetConfig(data json.RawMessage) {
	p.config = data
}

func (p *pluginImpl) GetAccessMethodDescriptor(name, version string) *internal.AccessMethodDescriptor {
	if !p.IsValid() {
		return nil
	}

	var fallback internal.AccessMethodDescriptor
	fallbackFound := false
	for _, m := range p.descriptor.AccessMethods {
		if m.Name == name {
			if m.Version == version {
				return &m
			}
			if m.Version == "" || m.Version == "v1" {
				fallback = m
				fallbackFound = true
			}
		}
	}
	if fallbackFound && (version == "" || version == "v1") {
		return &fallback
	}
	return nil
}

func (p *pluginImpl) Message() string {
	if p.IsValid() {
		return p.descriptor.Short
	}
	if p.error != "" {
		return "Error: " + p.error
	}
	return "unknown state"
}

func (p *pluginImpl) Validate(spec []byte) (*ppi.AccessSpecInfo, error) {
	result, err := p.Exec(nil, nil, "accessmethod", "validate", string(spec))
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

func (p *pluginImpl) Exec(r io.Reader, w io.Writer, args ...string) ([]byte, error) {
	return Exec(p.path, p.config, r, w, args...)
}

func Exec(execpath string, config json.RawMessage, r io.Reader, w io.Writer, args ...string) ([]byte, error) {
	if len(config) > 0 {
		args = append([]string{"-c", string(config)}, args...)
	}
	cmd := exec.Command(execpath, args...)

	stdout := w
	if w == nil {
		stdout = LimitBuffer(LIMIT)
	}

	stderr := LimitBuffer(LIMIT)

	cmd.Stdin = r
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	if err != nil {
		var result cmds.Error
		var msg string
		if err := json.Unmarshal(stderr.Bytes(), &result); err == nil {
			msg = result.Error
		} else {
			msg = fmt.Sprintf("[%s]", string(stderr.Bytes()))
		}
		return nil, fmt.Errorf("%s", msg)
	}
	if l, ok := stdout.(*LimitedBuffer); ok {
		if l.Exceeded() {
			return nil, fmt.Errorf("stdout limit exceeded")
		}
		return l.Bytes(), nil
	}
	return nil, nil
}
