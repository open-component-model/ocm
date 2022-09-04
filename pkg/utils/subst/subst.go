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

package subst

import (
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type SubstitutionTarget interface {
	SubstituteByData(path string, value []byte) error
	SubstituteByValue(path string, value interface{}) error

	Content() ([]byte, error)
}

func ParseFile(file string, fss ...vfs.FileSystem) (SubstitutionTarget, error) {
	fs := accessio.FileSystem(fss...)

	data, err := vfs.ReadFile(fs, file)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read file %q", file)
	}
	s, err := Parse(data)
	if err != nil {
		return nil, errors.Wrapf(err, "file %q", file)
	}
	return s, nil
}

func Parse(data []byte) (SubstitutionTarget, error) {
	var (
		err     error
		content interface{}
		fi      fileinfo
	)

	fi.json = true
	if err = runtime.DefaultJSONEncoding.Unmarshal(data, &content); err != nil {
		fi.json = false
		if err = runtime.DefaultYAMLEncoding.Unmarshal(data, &content); err != nil {
			return nil, errors.Wrapf(err, "no yaml or json data")
		}
		data, err = runtime.DefaultYAMLEncoding.Marshal(content)
	} else {
		data, err = runtime.DefaultJSONEncoding.Marshal(content)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "cannor marshal data")
	}
	// mixed json/yaml cannot be parsed, modified and marshalled again, correctly
	// so try to come with pure yaml or pure json.

	fi.content, err = parser.ParseBytes(data, 0)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid YAML")
	}
	return &fi, nil
}

type fileinfo struct {
	content *ast.File
	json    bool
}

func (f *fileinfo) Content() ([]byte, error) {
	data := []byte(f.content.String())

	if f.json {
		// TODO: the package seems to keep the file type json/yaml, but I'm not sure
		var err error
		data, err = yaml.YAMLToJSON([]byte(data))
		if err != nil {
			return nil, errors.Wrapf(err, "cannot marshal json")
		}
	}
	return data, nil
}

func (f *fileinfo) SubstituteByData(path string, value []byte) error {
	var m interface{}
	err := runtime.DefaultYAMLEncoding.Unmarshal(value, &m)
	if err != nil {
		return err
	}
	if f.json {
		value, err = runtime.DefaultJSONEncoding.Marshal(m)
	} else {
		value, err = runtime.DefaultYAMLEncoding.Marshal(m)
	}
	if err != nil {
		return err
	}
	return f.substituteByData(path, value)
}

func (f *fileinfo) substituteByData(path string, value []byte) error {
	file, err := parser.ParseBytes(value, 0)
	if err != nil {
		return errors.Wrapf(err, "cannot unmarshal value")
	}

	p, err := yaml.PathString("$." + path)
	if err != nil {
		return errors.Wrapf(err, "invalid substitution path")
	}
	return p.ReplaceWithFile(f.content, file)
}

func (f *fileinfo) SubstituteByValue(path string, value interface{}) error {
	var (
		err  error
		data []byte
	)
	if f.json {
		data, err = runtime.DefaultJSONEncoding.Marshal(value)
	} else {
		data, err = runtime.DefaultYAMLEncoding.Marshal(value)
	}
	if err != nil {
		return err
	}
	return f.substituteByData(path, data)
	/*
		node, err := yaml.ValueToNode(value)
		if err != nil {
			return errors.Wrapf(err, "cannot unmarshal value")
		}

		p, err := yaml.PathString("$." + path)
		if err != nil {
			return errors.Wrapf(err, "invalid substitution path")
		}
		return p.ReplaceWithNode(f.content, node)
	*/
}
