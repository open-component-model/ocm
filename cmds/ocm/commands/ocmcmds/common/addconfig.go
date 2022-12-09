// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"encoding/json"
	"fmt"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/pflag"
	"sigs.k8s.io/yaml"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/addhdlrs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/template"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/errors"
)

type ModifiedResourceSpecificationsFile struct {
	ElementFileSource
	modified string
}

func NewModifiedResourceSpecificationsFile(data string, path string, fss ...vfs.FileSystem) addhdlrs.ElementSource {
	return &ModifiedResourceSpecificationsFile{
		ElementFileSource: ElementFileSource{
			filesystem: accessio.FileSystem(fss...),
			path:       addhdlrs.NewSourceInfo(path),
		},
		modified: data,
	}
}

func (r *ModifiedResourceSpecificationsFile) Get() (string, error) {
	return r.modified, nil
}

////////////////////////////////////////////////////////////////////////////////

type ResourceConfigAdderCommand struct {
	utils.BaseCommand

	Templating template.Options
	Adder      ElementSpecificationsProvider

	ConfigFile string
	Resources  []addhdlrs.ElementSource
	Envs       []string
}

func (o *ResourceConfigAdderCommand) AddFlags(fs *pflag.FlagSet) {
	fs.StringArrayVarP(&o.Envs, "settings", "s", nil, "settings file with variable settings (yaml)")
	o.Templating.AddFlags(fs)
	if o.Adder != nil {
		o.Adder.AddFlags(fs)
	}
}

func (o *ResourceConfigAdderCommand) Complete(args []string) error {
	o.ConfigFile = args[0]
	o.Templating.Complete(o.Context.FileSystem())

	if o.Adder != nil {
		err := o.Adder.Complete()
		if err != nil {
			return err
		}

		rsc, err := o.Adder.Resources()
		if err != nil {
			return err
		}
		o.Resources = append(o.Resources, rsc...)
	}

	err := o.Templating.ParseSettings(o.Context.FileSystem(), o.Envs...)
	if err != nil {
		return err
	}

	paths := o.Templating.FilterSettings(args[1:]...)
	for _, p := range paths {
		o.Resources = append(o.Resources, NewElementFileSource(p, o.FileSystem()))
	}

	if len(o.Resources) == 0 {
		return fmt.Errorf("no specifications given")
	}
	return nil
}

func (o *ResourceConfigAdderCommand) ProcessResourceDescriptions(h ResourceSpecHandler) error {
	fs := o.Context.FileSystem()
	ictx := inputs.NewContext(o.Context, common.NewPrinter(o.Context.StdOut()), o.Templating.Vars)
	mode := vfs.FileMode(0o600)
	listkey := utils.Plural(h.Key(), 0)

	var current string
	if ok, err := vfs.FileExists(fs, o.ConfigFile); ok {
		fi, err := fs.Stat(o.ConfigFile)
		if err != nil {
			return errors.Wrapf(err, "cannot stat %s config file %q", listkey, o.ConfigFile)
		}
		mode = fi.Mode().Perm()
		data, err := vfs.ReadFile(fs, o.ConfigFile)
		if err != nil {
			return errors.Wrapf(err, "cannot read %s config file %q", listkey, o.ConfigFile)
		}
		current = string(data)
	} else if err != nil {
		return errors.Wrapf(err, "cannot read %s config file %q", listkey, o.ConfigFile)
	}

	for _, source := range o.Resources {
		r, err := source.Get()
		if err != nil {
			return err
		}
		var tmp map[string]interface{}
		err = json.Unmarshal([]byte(r), &tmp)
		if err == nil {
			b, err := yaml.Marshal(tmp)
			if err != nil {
				return errors.Wrapf(err, "cannot convert to YAML")
			}
			r = string(b)
		}
		current += "\n---\n" + r
	}

	source := NewModifiedResourceSpecificationsFile(current, o.ConfigFile, fs)
	resources, err := addhdlrs.DetermineElementsForSource(o.Context, ictx, o.Templating, h, source)
	if err != nil {
		return errors.Wrapf(err, "%s", source.Origin())
	}

	ictx.Printf("found %d %s\n", len(resources), listkey)

	err = vfs.WriteFile(fs, o.ConfigFile, []byte(current), mode)
	if err != nil {
		return errors.Wrapf(err, "cannot write %s config file %q", listkey, o.ConfigFile)
	}

	return nil
}
