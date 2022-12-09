// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"fmt"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/formatoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/addhdlrs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/addhdlrs/comp"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/fileoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/schemaoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/template"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	common2 "github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/errors"
)

var (
	Names = names.Components
	Verb  = verbs.Add
)

type Command struct {
	utils.BaseCommand

	Force  bool
	Create bool

	Handler ctf.FormatHandler
	Format  string

	Version    string
	Templating template.Options
	Envs       []string

	Archive string

	Elements []addhdlrs.ElementSource
}

func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx, formatoption.New(ctf.GetFormats()...), fileoption.New("transport-archive"), schemaoption.New(compdesc.DefaultSchemeVersion), &template.Options{})}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] [--version <version>] [<ctf archive>] {<components.yaml>}",
		Args:  cobra.MinimumNArgs(1),
		Short: "add component version(s) to a (new) transport archive",
		Example: `
<pre>
$ ocm add componentversions --file ctf --version 1.0 components.yaml
</pre>

and a file <code>components.yaml</code>:

<pre>
name: ocm.software/demo/test
version: 1.0.0
provider:
  name: ocm.software
  labels:
    - name: city
      value: Karlsruhe
labels:
  - name: purpose
    value: test

resources:
  - name: text
    type: PlainText
    input:
      type: file
      path: testdata
  - name: data
    type: PlainText
    input:
      type: binary
      data: IXN0cmluZ2RhdGE=

</pre>

The resource <code>text</code> is taken from a file <code>testdata</code> located
next to the description file.
`,
		Long: `
Add component versions specified by a description file to a Common Transport
Archive. This might be either a directory prepared to host component version
content or a tar/tgz file (see option --type).

If option <code>--create</code> is given, the archive is created first. An
additional option <code>--force</code> will recreate an empty archive if it already exists.

The source, resource and reference list can be composed according the commands
<CMD>ocm add sources</CMD>, <CMD>ocm add resources</CMD>, <CMD>ocm add references</CMD>, respectively.

The description file might contain:
- a single component as shown in the example
- a list of components under the key <code>components</code>
- a list of yaml documents with a single component or component list
`,
	}
}

func (o *Command) AddFlags(fs *pflag.FlagSet) {
	o.BaseCommand.AddFlags(fs)
	fs.BoolVarP(&o.Force, "force", "f", false, "remove existing content")
	fs.BoolVarP(&o.Create, "create", "c", false, "(re)create archive")
	fs.StringArrayVarP(&o.Envs, "settings", "s", nil, "settings file with variable settings (yaml)")
	fs.StringVarP(&o.Version, "version", "v", "", "default version for components")
}

func (o *Command) Complete(args []string) error {
	err := o.OptionSet.ProcessOnOptions(options.CompleteOptionsWithCLIContext(o.Context))
	if err != nil {
		return err
	}

	o.Archive, args = fileoption.From(o).GetPath(args, o.Context.FileSystem())
	o.Templating.Complete(o.Context.FileSystem())

	err = o.Templating.ParseSettings(o.Context.FileSystem(), o.Envs...)
	if err != nil {
		return err
	}

	paths := o.Templating.FilterSettings(args...)
	for _, p := range paths {
		o.Elements = append(o.Elements, common.NewElementFileSource(p, o.FileSystem()))
	}

	if len(o.Elements) == 0 {
		return fmt.Errorf("no specifications given")
	}

	format := formatoption.From(o).Format
	o.Handler = ctf.GetFormat(format)
	if o.Handler == nil {
		return accessio.ErrInvalidFileFormat(format.String())
	}

	return nil
}

func (o *Command) Run() error {
	var err error

	mode := formatoption.From(o).Mode()
	fs := o.Context.FileSystem()
	fp := fileoption.From(o).Path
	if ok, err := vfs.Exists(fs, fp); ok || err != nil {
		if err != nil {
			return err
		}
		if o.Force {
			err = fs.RemoveAll(fp)
			if err != nil {
				return errors.Wrapf(err, "cannot remove old %q", fp)
			}
		}
	}
	var repo ocm.Repository
	if o.Create {
		repo, err = ctf.Create(o.Context.OCMContext(), accessobj.ACC_CREATE, fp, mode, o.Handler, fs)
	} else {
		repo, err = ctf.Open(o.Context.OCMContext(), accessobj.ACC_CREATE, fp, mode, fs)
	}
	if err == nil {
		err = comp.ProcessComponentDescriptions(o.Context, common2.NewPrinter(o.Context.StdOut()), o.Templating, repo, comp.NewResourceSpecHandler(o.Version), o.Elements)
		cerr := repo.Close()
		if err == nil {
			err = cerr
		}
	}
	if err != nil {
		if o.Create {
			fs.RemoveAll(fp)
		}
		return err
	}

	return err
}
