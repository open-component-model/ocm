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

package add

import (
	"bytes"
	"fmt"
	"io"

	"github.com/gardener/ocm/cmds/ocm/cmd"
	"github.com/gardener/ocm/cmds/ocm/pkg/template"
	"github.com/gardener/ocm/cmds/ocm/pkg/utils"
	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/common/accessobj"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/ocm/compdesc"
	metav1 "github.com/gardener/ocm/pkg/ocm/compdesc/meta/v1"
	compdescv2 "github.com/gardener/ocm/pkg/ocm/compdesc/versions/v2"
	"github.com/gardener/ocm/pkg/ocm/repositories/ctf/comparch"
	"github.com/gardener/ocm/pkg/runtime"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

type Command struct {
	Context cmd.Context

	Archive    string
	Paths      []string
	Envs       []string
	Templating *template.Options
}

// NewCommand creates a new ctf command.
func NewCommand(ctx cmd.Context) *cobra.Command {
	return utils.SetupCommand(&Command{Context: ctx},
		&cobra.Command{
			Use:     "resources [<options>] <target> {<resourcefile> | <var>=<value>}",
			Args:    cobra.MinimumNArgs(2),
			Aliases: []string{"res", "resource"},
			Short:   "add resources to a component version",
			Long: `
Add resources specified in a resource file to a component version.
So far only component archives are supported as target.
`,
		})
}

func (o *Command) AddFlags(fs *pflag.FlagSet) {
	fs.StringArrayVarP(&o.Envs, "env", "e", nil, "environment file with variable settings")
}

func (o *Command) Complete(args []string) error {
	o.Archive = args[0]
	o.Templating = template.NewTemplateOptions()

	err := o.Templating.ParseSettings(o.Context.FileSystem(), o.Envs...)
	if err != nil {
		return err
	}

	o.Paths = o.Templating.FilterSettings(args[1:]...)

	return nil
}

type Resource struct {
	path   string
	source string
	spec   *ResourceOptions
}

func NewResource(spec *ResourceOptions, path string, indices ...int) *Resource {
	id := path
	for _, i := range indices {
		id += fmt.Sprintf("[%d]", i)
	}
	return &Resource{
		path:   path,
		source: id,
		spec:   spec,
	}
}

func (o *Command) Run() error {
	fs := o.Context.FileSystem()

	resources := []*Resource{}

	for _, filePath := range o.Paths {
		data, err := vfs.ReadFile(fs, filePath)
		//data, err := ioutil.ReadFile(p)
		if err != nil {
			return errors.Wrapf(err, "cannot read resource file %q", filePath)
		}

		parsed, err := o.Templating.Template(string(data))
		if err != nil {
			return errors.Wrapf(err, "error during variable substitution for %q", filePath)
		}
		// sigs parser has no multi document stream parsing
		// but yaml.v3 does not recognize json tagged fields.
		// Therefore we first use the v3 parser to parse the multi doc,
		// marshal it again and finally unmarshal it with the sigs parser.
		decoder := yaml.NewDecoder(bytes.NewBuffer([]byte(parsed)))
		i := 0
		for {
			var tmp interface{}
			desc := &Resources{}
			i++
			err := decoder.Decode(&tmp)
			if err != nil {
				if err != io.EOF {
					return err
				}
				break
			}
			data, err := yaml.Marshal(tmp)
			if err != nil {
				return err
			}
			err = runtime.DefaultYAMLEncoding.Unmarshal(data, desc)
			if err != nil {
				return err
			}
			if desc.ResourceOptions != nil {
				if desc.ResourceOptionList != nil {
					return errors.Newf("invalid resource spec %d in %q: either a list or a single resource possible", i+1, filePath)
				}
				if err = Validate(desc.ResourceOptions, o.Context, filePath); err != nil {
					return errors.Wrapf(err, "invalid resource spec %d in %q", i+1, filePath)
				}
				resources = append(resources, NewResource(desc.ResourceOptions, filePath, i))
			} else {
				if desc.ResourceOptionList == nil {
					return errors.Newf("invalid resource spec %d in %q: either a list or a single resource must be specified", i+1, filePath)
				}
				for j, r := range desc.ResourceOptionList.Resources {
					if err = Validate(r, o.Context, filePath); err != nil {
						return errors.Wrapf(err, "invalid resource spec %d[%d] in %q", i+1, j+1, filePath)
					}
					resources = append(resources, NewResource(r, filePath, i, j))
				}
			}
		}
	}

	obj, err := comparch.Open(o.Context.OCMContext(), accessobj.ACC_WRITABLE, o.Archive, 0, accessio.PathFileSystem(fs))
	if err != nil {
		return err
	}
	defer obj.Close()

	for _, r := range resources {
		vers := r.spec.Version
		if r.spec.Relation == metav1.LocalRelation {

			if vers == "" || vers == "<componentversion>" {
				vers = obj.GetVersion()
			} else {
				if vers != obj.GetVersion() {
					return errors.Newf("local resource %q (%s) has non-matching version %q", r.spec.Name, r.source, vers)
				}
			}
		}
		meta := &compdesc.ResourceMeta{
			ElementMeta: compdesc.ElementMeta{
				Name:          r.spec.Name,
				Version:       vers,
				ExtraIdentity: r.spec.ExtraIdentity,
				Labels:        r.spec.Labels,
			},
			Type:      r.spec.Type,
			Relation:  r.spec.Relation,
			SourceRef: compdescv2.Convert_SourceRefs_to(r.spec.SourceRef),
		}
		if r.spec.Input != nil {
			// Local Blob
			blob, hint, err := r.spec.Input.GetBlob(fs, r.path)
			if err != nil {
				return errors.Wrapf(err, "cannot get resource blob for %q(%s)", r.spec.Name, r.source)
			}
			err = obj.AddResourceBlob(meta, blob, hint, nil)
		} else {
			compdesc.GenericAccessSpec(r.spec.Access)
			err = obj.AddResource(meta, compdesc.GenericAccessSpec(r.spec.Access))
		}
		if err != nil {
			return errors.Wrapf(err, "cannot add resource %q(%s)", r.spec.Name, r.source)
		}
	}
	return nil
}
