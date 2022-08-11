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

package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/contexts/config"
	"github.com/open-component-model/ocm/pkg/errors"

	_ "github.com/open-component-model/ocm/pkg/contexts/datacontext/config"
)

func Configure(file string) error {
	h := os.Getenv("HOME")
	if file == "" {
		if h != "" {
			cfg := h + "/.ocmconfig"
			if ok, err := vfs.FileExists(osfs.New(), cfg); ok && err == nil {
				file = cfg
			}
		}
	}

	if file != "" {
		if strings.HasPrefix(file, "~"+string(os.PathSeparator)) {
			if len(h) == 0 {
				return fmt.Errorf("no home directory found for resolving path of config file %q", file)
			}
			file = h + file[1:]
		}
		data, err := os.ReadFile(file)
		if err != nil {
			return errors.Wrapf(err, "cannot read config file %q", file)
		}

		cfg, err := config.DefaultContext().GetConfigForData(data, nil)
		if err != nil {
			return errors.Wrapf(err, "invalid config file %q", file)
		}
		err = config.DefaultContext().ApplyConfig(cfg, file)
		if err != nil {
			return errors.Wrapf(err, "cannot apply config %q", file)
		}
	}
	return nil
}
