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

package localize

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/runtime"

	"github.com/open-component-model/ocm/pkg/errors"
)

type fileinfo struct {
	content interface{}
	json    bool
}

func Substitute(subs Substitutions, fs vfs.FileSystem) error {
	files := map[string]fileinfo{}

	for i, s := range subs {
		file, err := vfs.Canonical(fs, s.FilePath, true)
		if err != nil {
			return errors.Wrapf(err, "entry %d", i)
		}

		fi, ok := files[file]
		if !ok {
			data, err := vfs.ReadFile(fs, file)
			if err != nil {
				return errors.Wrapf(err, "entry %d: cannot read file %q", i, file)
			}
			fi.json = true
			if err = runtime.DefaultJSONEncoding.Unmarshal(data, &fi.content); err != nil {
				if err = runtime.DefaultYAMLEncoding.Unmarshal(data, &fi.content); err != nil {
					return errors.Wrapf(err, "entry %d: invalid YAML file %q", i, file)
				}
				fi.json = false
			}
			files[file] = fi
		}

		value, err := s.GetValue()
		if err != nil {
			return errors.Wrapf(err, "entry %d: cannot unmarshal value", i+1)
		}
		err = Set(fi.content, s.ValuePath, value)
		if err != nil {
			return errors.Wrapf(err, "entry %d: cannot substitute value", i+1)
		}
	}

	for file, fi := range files {
		marshal := runtime.DefaultYAMLEncoding.Marshal
		if fi.json {
			marshal = runtime.DefaultJSONEncoding.Marshal
		}

		data, err := marshal(fi.content)
		if err != nil {
			return errors.Wrapf(err, "cannot marshal %q after substitution ", file)
		}

		err = vfs.WriteFile(fs, file, data, vfs.ModePerm)
		if err != nil {
			return errors.Wrapf(err, "file %q", file)
		}
	}
	return nil
}

func Set(content interface{}, path string, value interface{}) error {
	values, ok := content.(map[string]interface{})
	if !ok {
		return fmt.Errorf("content must be a map")
	}
	fields := strings.Split(path, ".")
	i := 0
	for ; i < len(fields)-1; i++ {
		f := strings.TrimSpace(fields[i])
		v, ok := values[f]
		if !ok {
			v = map[string]interface{}{}
			values[f] = v
		} else {
			if _, ok := v.(map[string]interface{}); !ok {
				return fmt.Errorf("invalid field path %s", strings.Join(fields[:i+1], "."))
			}
		}
		values = v.(map[string]interface{})
	}
	values[fields[len(fields)-1]] = value
	return nil
}
