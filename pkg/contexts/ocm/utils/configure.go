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

package utils

import (
	"fmt"
	"os"
	"strings"

	dockercli "github.com/docker/cli/cli/config"
	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/spiff/features"
	"github.com/mandelsoft/spiff/spiffing"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common/accessio"

	"github.com/open-component-model/ocm/pkg/contexts/config"
	credcfg "github.com/open-component-model/ocm/pkg/contexts/credentials/config"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/dockerconfig"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/errors"
)

const DEFAULT_OCM_CONFIG = ".ocmconfig"

func Configure(ctx ocm.Context, path string, fss ...vfs.FileSystem) (ocm.Context, error) {
	fs := accessio.FileSystem(fss...)
	if ctx == nil {
		ctx = ocm.DefaultContext()
	}
	h := os.Getenv("HOME")
	if path == "" {
		if h != "" {
			cfg := h + "/" + DEFAULT_OCM_CONFIG
			if ok, err := vfs.FileExists(fs, cfg); ok && err == nil {
				path = cfg
			} else {
				cfg := h + "/ocm/ocmconfig"
				if ok, err := vfs.FileExists(fs, cfg); ok && err == nil {
					path = cfg
				}
			}
		}
	}
	if path != "" {
		if strings.HasPrefix(path, "~"+string(os.PathSeparator)) {
			if len(h) == 0 {
				return nil, fmt.Errorf("no home directory found for resolving path of ocm config file %q", path)
			}
			path = h + path[1:]
		}
		data, err := vfs.ReadFile(fs, path)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot read ocm config file %q", path)
		}

		sctx := spiffing.New().WithFeatures(features.INTERPOLATION, features.CONTROL)
		data, err = spiffing.Process(sctx, spiffing.NewSourceData(path, data))
		if err != nil {
			return nil, errors.Wrapf(err, "processing ocm config %q", path)
		}
		cfg, err := config.DefaultContext().GetConfigForData(data, nil)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid ocm config file %q", path)
		}
		err = config.DefaultContext().ApplyConfig(cfg, path)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot apply ocm config %q", path)
		}
	} else {
		// use docker config as default config for ocm cli
		d := filepath.Join(dockercli.Dir(), dockercli.ConfigFileName)
		if ok, err := vfs.FileExists(osfs.New(), d); ok && err == nil {
			cfg := credcfg.New()
			cfg.AddRepository(dockerconfig.NewRepositorySpec(d, true))
			err = config.DefaultContext().ApplyConfig(cfg, d)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot apply docker config %q", d)
			}
		}
	}
	return ctx, nil
}
