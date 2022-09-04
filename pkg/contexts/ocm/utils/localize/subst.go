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
	yaml "github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/utils/subst"

	"github.com/open-component-model/ocm/pkg/errors"
)

func Substitute(subs Substitutions, fs vfs.FileSystem) error {
	files := map[string]subst.SubstitutionTarget{}

	for i, s := range subs {
		file, err := vfs.Canonical(fs, s.FilePath, true)
		if err != nil {
			return errors.Wrapf(err, "entry %d", i)
		}

		fi, ok := files[file]
		if !ok {
			s, err := subst.ParseFile(file, fs)
			if err != nil {
				return errors.Wrapf(err, "entry %d", i)
			}
			files[file], fi = s, s
		}

		if err = fi.SubstituteByData(s.ValuePath, s.Value); err != nil {
			return errors.Wrapf(err, "entry %d: cannot substitute value", i+1)
		}
	}

	for file, fi := range files {
		data, err := fi.Content()
		if err != nil {
			return errors.Wrapf(err, "cannot marshal %q after substitution ", file)
		}

		if err := vfs.WriteFile(fs, file, data, vfs.ModePerm); err != nil {
			return errors.Wrapf(err, "file %q", file)
		}
	}
	return nil
}

func Set(content *ast.File, path string, value *ast.File) error {
	p, err := yaml.PathString("$." + path)
	if err != nil {
		return errors.Wrapf(err, "invalid substitution path")
	}
	return p.ReplaceWithFile(content, value)
}
