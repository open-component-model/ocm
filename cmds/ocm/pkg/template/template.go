// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"github.com/drone/envsubst"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

// Options defines the options for component-cli templating
type Options struct {
	Vars map[string]string
}

// NewTemplateOptions provides a new templating options object
func NewTemplateOptions() *Options {
	return &Options{map[string]string{}}
}

// Usage prints out the usage for templating
func (o *Options) Usage() string {
	return `
Templating:
All yaml/json defined resources can be templated using simple envsubst syntax.
Variables are specified as regular arguments following the syntax "<name>=<value>".

Note: Variable names are case-sensitive.

Example:
<pre>
<command> <options> -- MY_VAL=test <args>
</pre>

<pre>

key:
  subkey: "abc ${MY_VAL}"

</pre>

`
}

// FilterSettings parses commandline argument variables.
// it returns all non variable arguments
func (o *Options) FilterSettings(args ...string) []string {
	var addArgs []string
	if o.Vars == nil {
		o.Vars = map[string]string{}
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
		o.Vars = map[string]string{}
	}
	for _, path := range paths {
		vars, err := ReadSettings(fs, path)
		if err != nil {
			return errors.Wrapf(err, "cannot read env file %q", path)
		}
		for k, v := range vars {
			o.Vars[k] = v
		}
	}
	return nil
}

// Template templates a string with the parsed vars.
func (o *Options) Template(data string) (string, error) {
	return envsubst.Eval(data, o.mapping)
}

// mapping is a helper function for the envsubst to provide the value for a variable name.
// It returns an emtpy string if the variable is not defined.
func (o *Options) mapping(variable string) string {
	if o.Vars == nil {
		return ""
	}
	// todo: maybe use os.getenv as backup.
	return o.Vars[variable]
}

func ReadSettings(fs vfs.FileSystem, path string) (map[string]string, error) {
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
	if err == io.EOF {
		err = nil
	}
	return result, err
}
