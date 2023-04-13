// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package helm

import (
	"fmt"

	"github.com/mandelsoft/logging"

	"github.com/open-component-model/ocm/cmds/helminstaller/app/driver"
)

type Driver struct{}

var _ driver.Driver = Driver{}

func New() driver.Driver {
	return Driver{}
}

func (Driver) Install(cfg *driver.Config) error {
	return Install(cfg.Debug, cfg.ChartPath, cfg.Release, cfg.Namespace, cfg.CreateNamespace, cfg.Values, cfg.Kubeconfig)
}

func (Driver) Uninstall(cfg *driver.Config) error {
	return Uninstall(cfg.Debug, cfg.ChartPath, cfg.Release, cfg.Namespace, cfg.CreateNamespace, cfg.Values, cfg.Kubeconfig)
}

func DebugFunction(l logging.Logger) func(msg string, args ...any) {
	return func(msg string, args ...any) {
		l.Debug(fmt.Sprintf(msg, args...))
	}
}
