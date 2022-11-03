// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"

	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type Values map[string]interface{}

type Templater interface {
	Process(data string, values Values) (string, error)
}

// Options defines the options for cli templating.
type Options struct {
	Default   string
	Mode      string
	UseEnv    bool
	Templater Templater
	Vars      Values
}

func (o *Options) defaultMode() string {
	if o.Default == "" {
		return "subst"
	}
	return o.Default
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Mode, "templater", "", o.defaultMode(), "templater to use (subst, spiff, go)")
	fs.BoolVarP(&o.UseEnv, "addenv", "", false, "access environment for templating")
}

func (o *Options) Complete(fs vfs.FileSystem) error {
	var err error

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
	o.Templater, err = DefaultRegistry().Create(o.Mode, fs)
	if err != nil {
		return err
	}
	return nil
}

// Usage prints out the usage for templating.
func (o *Options) Usage() string {
	return `
All yaml/json defined resources can be templated.
Variables are specified as regular arguments following the syntax <code>&lt;name>=&lt;value></code>.
Additionally settings can be specified by a yaml file using the <code>--settings <file></code>
option. With the option <code>--addenv</code> environment variables are added to the binding.
Values are overwritten in the order environment, settings file, command line settings. 

Note: Variable names are case-sensitive.

Example:
<pre>
&lt;command> &lt;options> -- MY_VAL=test &lt;args>
</pre>
` + Usage(DefaultRegistry())
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

func SplitYamlDocuments(data []byte) ([][]byte, error) {
	decoder := yaml.NewDecoder(bytes.NewBuffer([]byte(data)))
	list := [][]byte{}
	i := 0
	for {
		var tmp interface{}

		i++
		err := decoder.Decode(&tmp)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return nil, errors.Wrapf(err, "cannot parse document %d", i)
			}
			break
		}
		out, err := yaml.Marshal(tmp)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot marshal document %d", i)
		}

		list = append(list, out)
	}
	return list, nil
}
