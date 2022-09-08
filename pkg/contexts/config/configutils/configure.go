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

package configutils

import (
	"fmt"
	"os"
	"strings"

	_ "github.com/open-component-model/ocm/pkg/contexts/datacontext/config"

	"github.com/mandelsoft/spiff/features"
	"github.com/mandelsoft/spiff/spiffing"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/contexts/config"
	"github.com/open-component-model/ocm/pkg/errors"
)

func Configure(path string) error {
	h := os.Getenv("HOME")
	if path == "" {
		if h != "" {
			cfg := h + "/.ocmconfig"
			if ok, err := vfs.FileExists(osfs.New(), cfg); ok && err == nil {
				path = cfg
			}
		}
	}

	if path != "" {
		if strings.HasPrefix(path, "~"+string(os.PathSeparator)) {
			if len(h) == 0 {
				return fmt.Errorf("no home directory found for resolving path of ocm config file %q", path)
			}
			path = h + path[1:]
		}
		data, err := vfs.ReadFile(osfs.New(), path)
		if err != nil {
			return errors.Wrapf(err, "cannot read ocm config file %q", path)
		}

		sctx := spiffing.New().WithFeatures(features.INTERPOLATION, features.CONTROL)
		data, err = spiffing.Process(sctx, spiffing.NewSourceData(path, data))
		if err != nil {
			return errors.Wrapf(err, "processing ocm config %q", path)
		}
		cfg, err := config.DefaultContext().GetConfigForData(data, nil)
		if err != nil {
			return errors.Wrapf(err, "invalid ocm config file %q", path)
		}
		err = config.DefaultContext().ApplyConfig(cfg, path)
		if err != nil {
			return errors.Wrapf(err, "cannot apply ocm config %q", path)
		}
	}
	return nil
}
