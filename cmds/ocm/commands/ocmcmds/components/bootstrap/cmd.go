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

package bootstrap

import (
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/pkg/contexts/clictx"

	"github.com/open-component-model/ocm/pkg/common/accessio"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/install"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/install/drivers/docker"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/out"
	"github.com/open-component-model/ocm/pkg/runtime"

	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"

	ocmcommon "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/handlers/comphdlr"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"

	topicbootstrap "github.com/open-component-model/ocm/cmds/ocm/topics/ocm/bootstrapping"
)

var (
	Names = names.Components
	Verb  = verbs.Bootstrap
)

type Command struct {
	utils.BaseCommand
	Action string
	Ref    string
	Id     v1.Identity

	CredentialsFile string
	ParameterFile   string
	OutputFile      string
	Credentials     accessio.DataSource
	Parameters      accessio.DataSource
}

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx, repooption.New(), lookupoption.New())}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "[<options>] <action> {<component-reference>} {<resource id field>}",
		Args:  cobra.MinimumNArgs(2),
		Short: "bootstrap component version",
		Long: `
Use the simple OCM bootstrap mechanism to execute a bootstrap resource.

The bootstrap resource must have the type <code>` + install.TypeOCMInstaller + `</code>.
This is a simple YAML file resource describing the bootstrapping. See also the
topic <CMD>ocm ocm-bootstrapping</CMD>.

The first matching resource of this type is selected. Optionally a set of
identity attribute can be specified used to refine the match. This can be the
resource name and/or other key/value pairs (<code>&lt;attr>=&lt;value></code>).

If no output file is provided, the yaml representation of the outputs are
printed to standard out. If the output file is a directory, for every output a
dedicated file is created, otherwise the yaml representation is stored to the
file.
`,
		Example: `
$ ocm bootstrap componentversion ghcr.io/mandelsoft/ocmdemoinstaller:0.0.1-dev
`,
	}
	cmd.AddCommand(topicbootstrap.New(o.Context))
	return cmd
}

func (o *Command) AddFlags(set *pflag.FlagSet) {
	set.StringVarP(&o.CredentialsFile, "credentials", "c", "", "credentials file")
	set.StringVarP(&o.ParameterFile, "parameters", "p", "", "parameter file")
	set.StringVarP(&o.OutputFile, "outputs", "o", "", "output file/directory")
}

func (o *Command) Complete(args []string) error {
	o.Action = args[0]
	o.Ref = args[1]
	id, err := ocmcommon.MapArgsToIdentityPattern(args[2:]...)
	if err != nil {
		return errors.Wrapf(err, "bootstrap resource identity pattern")
	}
	o.Id = id
	if len(o.CredentialsFile) > 0 {
		data, err := vfs.ReadFile(o.Context.FileSystem(), o.CredentialsFile)
		if err != nil {
			return errors.Wrapf(err, "failed reading credentials file %q", o.CredentialsFile)
		}
		o.Credentials = accessio.DataAccessForBytes(data, o.CredentialsFile)
	}
	if len(o.ParameterFile) > 0 {
		data, err := vfs.ReadFile(o.Context.FileSystem(), o.ParameterFile)
		if err != nil {
			return errors.Wrapf(err, "failed reading parameter file %q", o.ParameterFile)
		}
		o.Parameters = accessio.DataAccessForBytes(data, o.ParameterFile)
	}
	return nil
}

func (o *Command) Run() error {
	session := ocm.NewSession(nil)
	defer session.Close()

	err := o.ProcessOnOptions(ocmcommon.CompleteOptionsWithSession(o, session))
	if err != nil {
		return err
	}
	handler := comphdlr.NewTypeHandler(o.Context.OCM(), session, repooption.From(o).Repository)
	return utils.HandleOutput(&action{cmd: o}, handler, utils.StringElemSpecs(o.Ref)...)
}

////////////////////////////////////////////////////////////////////////////////

type action struct {
	data comphdlr.Objects
	cmd  *Command
}

var _ output.Output = (*action)(nil)

func (a *action) Add(e interface{}) error {
	if len(a.data) > 0 {
		return errors.New("found multiple component versions")
	}
	a.data = append(a.data, e.(*comphdlr.Object))
	return nil
}

func (a *action) Close() error {
	return nil
}

type Binary struct {
	Binary []byte `json:"binary"`
}

func (a *action) Out() error {
	result, err := install.Install(&docker.Driver{}, a.cmd.Action, a.cmd.Id, a.cmd.Credentials, a.cmd.Parameters, a.cmd.OCMContext(), a.data[0].ComponentVersion, lookupoption.From(a.cmd))
	if err != nil {
		return err
	}

	if a.cmd.OutputFile != "" {
		if ok, _ := vfs.IsDir(a.cmd.FileSystem(), a.cmd.OutputFile); ok {
			out.Outf(a.cmd, "writing outputs to directory %q...", a.cmd.OutputFile)
			for n, o := range result.Outputs {
				err := vfs.WriteFile(a.cmd.FileSystem(), vfs.Join(a.cmd.FileSystem(), a.cmd.OutputFile, n), o, 0600)
				if err != nil {
					return errors.Wrapf(err, "cannot write output %q", n)
				}
			}
			return nil
		}
	}

	data := map[string]interface{}{}
	for n, o := range result.Outputs {
		var tmp interface{}
		err := runtime.DefaultYAMLEncoding.Unmarshal(o, &tmp)
		if err == nil {
			data[n] = tmp
		} else {
			data[n] = &Binary{o}
		}
	}

	outputs, err := runtime.DefaultYAMLEncoding.Marshal(map[string]interface{}{"outputs": data})
	if err != nil {
		return errors.Wrapf(err, "cannot marshal outputs")
	}
	if a.cmd.OutputFile != "" {
		vfs.WriteFile(a.cmd.FileSystem(), a.cmd.OutputFile, outputs, 0600)
	} else {
		out.Outf(a.cmd, "Provided outputs:\n%s\n", outputs)
	}
	return nil
}
