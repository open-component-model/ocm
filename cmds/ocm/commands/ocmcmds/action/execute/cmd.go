// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package execute

import (
	"encoding/json"
	"strings"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/action"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/out"
	"github.com/open-component-model/ocm/pkg/runtime"
)

var (
	Names = names.Action
	Verb  = verbs.Execute
)

type Command struct {
	utils.BaseCommand

	Name       string
	Spec       action.ActionSpec
	OutputMode string
}

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(
		&Command{
			BaseCommand: utils.NewBaseCommand(ctx),
		},
		utils.Names(Names, names...)...,
	)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] <action spec>",
		Short: "execute an action",
		Args:  cobra.ExactArgs(1),
		Long: `
Execute an action extension for a given action specification. The specification
show be a JSON or YAML argument.
`,
		Example: `
$ ocm execute action '{ "type": "oci.repository.prepare/v1", "hostname": "...", "repository": "..."}'
`,
	}
}

func (o *Command) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.OutputMode, "output", "o", "json", "output mode (json, yaml)")
	fs.StringVarP(&o.Name, "name", "n", "", "action name (overrides type in specification)")
}

func (o *Command) Complete(args []string) error {
	var err error

	data := []byte(args[0])
	if strings.HasPrefix(args[0], "@") {
		data, err = vfs.ReadFile(o.FileSystem(), args[0][1:])
		if err != nil {
			return errors.Wrapf(err, "cannot read file %q", args[0][1:])
		}
	}

	if o.OutputMode != "json" && o.OutputMode != "yaml" {
		return errors.Wrapf(err, "invalid output mode %q", o.OutputMode)
	}

	var un runtime.UnstructuredVersionedTypedObject

	err = runtime.DefaultYAMLEncoding.Unmarshal(data, &un)
	if err != nil {
		return errors.Wrapf(err, "invalid action spec")
	}
	if o.Name != "" {
		un.SetType(o.Name)
	} else {
		o.Name = un.GetKind()
	}

	data, err = json.Marshal(&un)
	if err != nil {
		return errors.Wrapf(err, "cannot marshal final spec")
	}
	o.Spec, err = action.DecodeActionSpec(data)
	return err
}

func (o *Command) Run() error {
	out.Outf(o, "Executing action %s...\n", o.Name)
	r, err := o.Context.OCMContext().GetActions().Execute(o.Spec, nil)
	if err != nil {
		return errors.Wrapf(err, "execution failed")
	}

	var data []byte

	if o.OutputMode == "json" {
		data, err = runtime.DefaultJSONEncoding.Marshal(r)
	} else {
		data, err = runtime.DefaultYAMLEncoding.Marshal(r)
	}
	if err != nil {
		return errors.Wrapf(err, "cannot marshal result")
	}
	out.Outf(o, "%s", string(data))
	return nil
}
