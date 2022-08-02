// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package scriptoption

import (
	"encoding/json"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	cfgcpi "github.com/open-component-model/ocm/pkg/contexts/config/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	ConfigType   = "scripts.ocm.config" + common.TypeGroupSuffix
	ConfigTypeV1 = ConfigType + runtime.VersionSeparator + "v1"
)

func init() {
	cfgcpi.RegisterConfigType(ConfigType, cfgcpi.NewConfigType(ConfigType, &Config{}, usage))
	cfgcpi.RegisterConfigType(ConfigTypeV1, cfgcpi.NewConfigType(ConfigTypeV1, &Config{}, usage))
}

// Config describes a memory based repository interface.
type Config struct {
	runtime.ObjectVersionedType `json:",inline"`
	Scripts                     map[string]ScriptSpec `json:"scripts"`
}

type ScriptSpec struct {
	Path       string          `json:"path,omitempty"`
	Script     json.RawMessage `json:"script,omitempty"`
	FileSystem vfs.FileSystem  `json:"-"`
}

// NewConfig creates a new memory ConfigSpec
func NewConfig() *Config {
	return &Config{
		ObjectVersionedType: runtime.NewVersionedObjectType(ConfigType),
	}
}

func (a *Config) GetType() string {
	return ConfigType
}

func (a *Config) AddScriptFile(name, path string, fss ...vfs.FileSystem) {
	var fs vfs.FileSystem
	for _, fs = range fss {
		if fs != nil {
			break
		}
	}
	if a.Scripts == nil {
		a.Scripts = map[string]ScriptSpec{}
	}
	a.Scripts[name] = ScriptSpec{Path: path, FileSystem: fs}
}

func (a *Config) AddScript(name string, data []byte) {
	if a.Scripts == nil {
		a.Scripts = map[string]ScriptSpec{}
	}
	a.Scripts[name] = ScriptSpec{Script: data}
}

func (a *Config) ApplyTo(ctx cfgcpi.Context, target interface{}) error {
	t, ok := target.(*Option)
	if !ok {
		return cfgcpi.ErrNoContext(ConfigType)
	}

	spec, ok := a.Scripts[t.Script]
	if ok {
		if len(spec.Script) > 0 {
			t.ScriptData = spec.Script
		} else {
			if spec.Path == "" {
				return errors.Newf("script or path must be set for entry %q", t.Script)
			}
			fs := accessio.FileSystem(spec.FileSystem, t.FileSystem)
			data, err := vfs.ReadFile(fs, spec.Path)
			if err != nil {
				return errors.Wrapf(err, "script file %q", spec.Path)
			}
			t.ScriptData = data
		}
	}
	return nil
}

const usage = `
The config type <code>` + ConfigType + `</code> can be used to define transfer scripts:

<pre>
    type: ` + ConfigType + `
    scripts:
      &lt;name>:
        path: &lt;>file path>
      &lt;other name>:
        script: &lt;>nested script as yaml>
</pre>
`
