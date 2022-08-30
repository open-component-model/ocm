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

package template

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type Values map[string]interface{}

type Templater interface {
	Process(data string, values Values) (string, error)
}

// Options defines the options for cli templating.
type Options struct {
	Mode      string
	UseEnv    bool
	Templater Templater
	Vars      Values
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Mode, "templater", "", "subst", "templater to use (subst, spiff, go)")
	fs.BoolVarP(&o.UseEnv, "addenv", "", false, "access environment for templating")
}

func (o *Options) Complete(fs vfs.FileSystem) error {
	if o.Vars == nil {
		o.Vars = Values{}
	}
	if o.Mode == "" {
		o.Mode = "subst"
	}
	if o.UseEnv {
		for _, v := range os.Environ() {
			if i := strings.Index(v, "="); i > 0 {
				value := v[i+1:]
				name := strings.TrimSpace(v[0:i])
				o.Vars[name] = value
			}
		}
	}
	switch o.Mode {
	case "subst":
		o.Templater = NewSubst()
	case "go":
		o.Templater = NewGo()
	case "spiff":
		o.Templater = NewSpiff(fs)
	default:
		return errors.Newf("unsupported templater %q", o.Mode)
	}
	return nil
}

// Usage prints out the usage for templating.
func (o *Options) Usage() string {
	return `
Templating:
All yaml/json defined resources can be templated.
Variables are specified as regular arguments following the syntax <code>&lt;name>=&lt;value></code>.
Additionally settings can be specified by a yaml file using the <code>--settings <file></code>
option. With the option <code>--addenv</code> environment variables are added to the binding.
Values are overwritten in the order environment, settings file, command line settings. 

Note: Variable names are case-sensitive.

Example:
<pre>
<command> <options> -- MY_VAL=test <args>
</pre>

There are several templaters that can be selected by the <code>--templater</code> option:
- envsubst: simple value substitution with the <code>drone/envsubst</code> templater. It
  supports string values, only. Complexity settings will be json encoded.
  <pre>
  key:
    subkey: "abc ${MY_VAL}"
  </pre>

- go: go templating supports complex values.
  <pre>
  key:
    subkey: "abc {{.MY_VAL}}"
  </pre>

- spiff: [spiff templating](https://github.com/mandelsoft/spiff) supports
  complex values. the settings are accessible using the binding <tt>values</tt>.
  <pre>
  key:
    subkey: "abc (( values.MY_VAL ))"
  </pre>
`
}

// FilterSettings parses commandline argument variables.
// it returns all non variable arguments.
func (o *Options) FilterSettings(args ...string) []string {
	var addArgs []string
	if o.Vars == nil {
		o.Vars = Values{}
	}
	for _, arg := range args {
		if i := strings.Index(arg, "="); i > 0 {
			value := arg[i+1:]
			name := strings.TrimSpace(arg[0:i])
			o.Vars[name] = value
			continue
		}
		addArgs = append(addArgs, arg)
	}
	return addArgs
}

func (o *Options) ParseSettings(fs vfs.FileSystem, paths ...string) error {
	if o.Vars == nil {
		o.Vars = Values{}
	}
	for _, path := range paths {
		vars, err := ReadYAMLSettings(fs, path)
		if err != nil {
			return errors.Wrapf(err, "cannot read env file %q", path)
		}
		for k, v := range vars {
			o.Vars[k] = v
		}
	}
	return nil
}

// Execute templates a string with the parsed vars.
func (o *Options) Execute(data string) (string, error) {
	return o.Templater.Process(data, o.Vars)
}

func ReadYAMLSettings(fs vfs.FileSystem, path string) (Values, error) {
	result := Values{}
	data, err := vfs.ReadFile(fs, path)
	if err != nil {
		return nil, err
	}
	err = runtime.DefaultYAMLEncoding.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func ReadSimpleSettings(fs vfs.FileSystem, path string) (map[string]string, error) {
	var (
		part   []byte
		prefix bool
	)

	result := map[string]string{}
	file, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	buffer := bytes.NewBuffer(make([]byte, 0, 1024))
	for {
		if part, prefix, err = reader.ReadLine(); err != nil {
			break
		}
		buffer.Write(part)
		if !prefix {
			line := strings.TrimSpace(buffer.String())
			if line != "" && !strings.HasPrefix(line, "#") {
				i := strings.Index(line, "=")
				if i <= 0 {
					return nil, errors.Newf("invalid variable syntax %q", line)
				}
				result[strings.TrimSpace(line[:i])] = strings.TrimSpace(line[i+1:])
			}
			buffer.Reset()
		}
	}
	if errors.Is(err, io.EOF) {
		err = nil
	}
	return result, err
}
