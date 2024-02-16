// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package dockerconfig

import (
	dockercli "github.com/docker/cli/cli/config"
	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/contexts/config"
	credcfg "github.com/open-component-model/ocm/pkg/contexts/credentials/config"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils/defaultconfigregistry"
	"github.com/open-component-model/ocm/pkg/errors"
)

func init() {
	defaultconfigregistry.RegisterDefaultConfigHandler(DefaultConfigHandler, desc)
}

func DefaultConfigHandler(cfg config.Context) error {
	// use docker config as default config for ocm cli
	d := filepath.Join(dockercli.Dir(), dockercli.ConfigFileName)
	if ok, err := vfs.FileExists(osfs.New(), d); ok && err == nil {
		ccfg := credcfg.New()
		ccfg.AddRepository(NewRepositorySpec(d, true))
		err = cfg.ApplyConfig(ccfg, d)
		if err != nil {
			return errors.Wrapf(err, "cannot apply docker config %q", d)
		}
	}
	return nil
}

var desc = `
The docker configuration file at <code>~/.docker/config.json</code> is
read to feed in the configured credentials for OCI registries.
`
