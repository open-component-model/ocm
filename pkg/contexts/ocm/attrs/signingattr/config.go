// Copyright 2020 Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package signingattr

import (
	"encoding/base64"
	"encoding/json"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/config"
	cfgcpi "github.com/open-component-model/ocm/pkg/contexts/config/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/signing"
)

const (
	ConfigType   = "keys.config" + common.TypeGroupSuffix
	ConfigTypeV1 = ConfigType + runtime.VersionSeparator + "v1"
)

func init() {
	cfgcpi.RegisterConfigType(ConfigType, cfgcpi.NewConfigType(ConfigType, &ConfigSpec{}, usage))
	cfgcpi.RegisterConfigType(ConfigTypeV1, cfgcpi.NewConfigType(ConfigTypeV1, &ConfigSpec{}, usage))
}

// ConfigSpec describes a memory based repository interface.
type ConfigSpec struct {
	runtime.ObjectVersionedType `json:",inline"`
	PublicKeys                  map[string]KeySpec `json:"publicKeys"`
	PrivateKeys                 map[string]KeySpec `json:"privateKeys"`
}

type RawData []byte

var _ json.Unmarshaler = (*RawData)(nil)

func (r RawData) MarshalJSON() ([]byte, error) {
	return json.Marshal(base64.StdEncoding.EncodeToString(r))
}
func (r *RawData) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	*r, err = base64.StdEncoding.DecodeString(s)
	return err
}

type KeySpec struct {
	Data       RawData        `json:"data,omitempty"`
	Path       string         `json:"path,omitempty"`
	Parsed     interface{}    `json:"-"`
	FileSystem vfs.FileSystem `json:"-"`
}

func (k *KeySpec) Get() (interface{}, error) {
	if k.Parsed != nil {
		return k.Parsed, nil
	}
	if k.Data != nil {
		return []byte(k.Data), nil
	}
	fs := k.FileSystem
	if fs == nil {
		fs = osfs.New()
	}
	return vfs.ReadFile(fs, k.Path)
}

// NewConfigSpec creates a new memory ConfigSpec
func NewConfigSpec() *ConfigSpec {
	return &ConfigSpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(ConfigType),
	}
}

func (a *ConfigSpec) GetType() string {
	return ConfigType
}

func (a *ConfigSpec) addKey(set *map[string]KeySpec, name string, key interface{}) {
	if *set == nil {
		*set = map[string]KeySpec{}
	}
	(*set)[name] = KeySpec{Parsed: key}
}

func (a *ConfigSpec) AddPublicKey(name string, key interface{}) {
	a.addKey(&a.PublicKeys, name, key)
}

func (a *ConfigSpec) AddPrivateKey(name string, key interface{}) {
	a.addKey(&a.PrivateKeys, name, key)
}

func (a *ConfigSpec) addKeyFile(set *map[string]KeySpec, name, path string, fss ...vfs.FileSystem) {
	var fs vfs.FileSystem
	for _, fs = range fss {
		if fs != nil {
			break
		}
	}
	if *set == nil {
		*set = map[string]KeySpec{}
	}
	(*set)[name] = KeySpec{Path: path, FileSystem: fs}
}

func (a *ConfigSpec) AddPublicKeyFile(name, path string, fss ...vfs.FileSystem) {
	a.addKeyFile(&a.PublicKeys, name, path, fss...)
}

func (a *ConfigSpec) AddPrivateKeyFile(name, path string, fss ...vfs.FileSystem) {
	a.addKeyFile(&a.PrivateKeys, name, path, fss...)
}

func (a *ConfigSpec) addKeyData(set *map[string]KeySpec, name string, data []byte) {
	if *set == nil {
		*set = map[string]KeySpec{}
	}
	(*set)[name] = KeySpec{Data: data}
}

func (a *ConfigSpec) AddPublicKeyData(name string, data []byte) {
	a.addKeyData(&a.PublicKeys, name, data)
}

func (a *ConfigSpec) AddPrivateKeyData(name string, data []byte) {
	a.addKeyData(&a.PrivateKeys, name, data)
}

func (a *ConfigSpec) ApplyTo(ctx cfgcpi.Context, target interface{}) error {
	t, ok := target.(config.Context)
	if !ok {
		return cfgcpi.ErrNoContext(ConfigType)
	}
	return errors.Wrapf(a.ApplyToRegistry(Get(t)), "applying config failed")
}

func (a *ConfigSpec) ApplyToRegistry(registry signing.KeyRegistry) error {
	for n, k := range a.PublicKeys {
		key, err := k.Get()
		if err != nil {
			return errors.Wrapf(err, "cannot get public key %s", n)
		}
		registry.RegisterPublicKey(n, key)
	}
	for n, k := range a.PrivateKeys {
		key, err := k.Get()
		if err != nil {
			return errors.Wrapf(err, "cannot get private key %s", n)
		}
		registry.RegisterPrivateKey(n, key)
	}
	return nil
}

const usage = `
The config type <code>` + ConfigType + `</code> can be used to define
public and private keys:

<pre>
    type: ` + ConfigType + `
    privateKeys:
       &lt;name>:
         path: &lt;file path>
       ...
    publicKeys:
       &lt;name>:
         data: &lt;base64 encoded key representation>
       ...
</pre>
`
