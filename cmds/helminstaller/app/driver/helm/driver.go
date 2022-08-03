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

package helm

import (
	"github.com/open-component-model/ocm/cmds/helminstaller/app/driver"
)

type Driver struct{}

var _ driver.Driver = Driver{}

func New() driver.Driver {
	return Driver{}
}

func (Driver) Install(cfg *driver.Config) error {
	return Install(cfg.ChartPath, cfg.Release, cfg.Namespace, cfg.CreateNamespace, cfg.Values, cfg.Kubeconfig)
}

func (Driver) Uninstall(cfg *driver.Config) error {
	return Uninstall(cfg.ChartPath, cfg.Release, cfg.Namespace, cfg.CreateNamespace, cfg.Values, cfg.Kubeconfig)
}
