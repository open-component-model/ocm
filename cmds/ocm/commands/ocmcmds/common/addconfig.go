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

	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/ocmcmds/common/addhdlrs"
	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/ocmcmds/common/options/templateroption"
	"github.com/open-component-model/ocm/v2/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/v2/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/v2/pkg/common"
	"github.com/open-component-model/ocm/v2/pkg/common/accessio"
	"github.com/open-component-model/ocm/v2/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/v2/pkg/errors"
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

	Adder ElementSpecificationsProvider

	ConfigFile string
	Resources  []addhdlrs.ElementSource
	Envs       []string
}

// NewCommand creates a new ctf command.
func NewResourceConfigAdderCommand(ctx clictx.Context, adder ElementSpecificationsProvider, opts ...options.Options) ResourceConfigAdderCommand {
	return ResourceConfigAdderCommand{
		BaseCommand: utils.NewBaseCommand(ctx, append(opts, templateroption.New("none"))...),
		Adder:       adder,
	}
}

func (o *ResourceConfigAdderCommand) AddFlags(fs *pflag.FlagSet) {
	fs.StringArrayVarP(&o.Envs, "settings", "s", nil, "settings file with variable settings (yaml)")
	if o.Adder != nil {
		o.Adder.AddFlags(fs)
	}
}

func (o *ResourceConfigAdderCommand) Complete(args []string) error {
	o.ConfigFile = args[0]

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

	t := templateroption.From(o)
	err := t.ParseSettings(o.Context.FileSystem(), o.Envs...)
	if err != nil {
		return err
	}

	paths := t.FilterSettings(args[1:]...)
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
	ictx := inputs.NewContext(o.Context, common.NewPrinter(o.Context.StdOut()), templateroption.From(o).Vars)
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
	resources, err := addhdlrs.DetermineElementsForSource(o.Context, ictx, templateroption.From(o).Options, h, source)
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
