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

package app

import (
	"github.com/open-component-model/ocm/cmds/helminstaller/app/driver"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/toi/support"
)

func New(d driver.Driver) func(o *support.ExecutorOptions) error {
	return func(o *support.ExecutorOptions) error {
		return Executor(d, o)
	}
}

func Executor(d driver.Driver, o *support.ExecutorOptions) error {
	var cfg Config
	err := runtime.DefaultYAMLEncoding.Unmarshal(o.ConfigData, &cfg)
	if err != nil {
		return errors.Wrapf(err, "cannot unmarshal config")
	}
	var values map[string]interface{}
	err = runtime.DefaultYAMLEncoding.Unmarshal(o.ParameterData, &values)
	if err != nil {
		return errors.Wrapf(err, "cannot unmarshal parameters")
	}
	if cfg.KubeConfigName == "" {
		cfg.KubeConfigName = "target"
	}
	creds, err := o.CredentialRepo.LookupCredentials(cfg.KubeConfigName)
	if err != nil {
		return errors.Wrapf(err, "cannot get kubeconfig with key %q", cfg.KubeConfigName)
	}
	v := creds.GetProperty("KUBECONFIG")
	if v == "" {
		return errors.Wrapf(err, "property KUBECONFIG missing in credential %q", cfg.KubeConfigName)
	}
	return Execute(d, o.Action, o.Context, o.OutputContext, o.ComponentVersion, &cfg, values, []byte(v))
}
