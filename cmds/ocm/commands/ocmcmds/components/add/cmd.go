package add

import (
	"fmt"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/mandelsoft/goutils/errors"
	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/formatoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/addhdlrs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/addhdlrs/comp"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/dryrunoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/fileoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/rscbyvalueoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/schemaoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/templateroption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	topicocmlabels "github.com/open-component-model/ocm/cmds/ocm/topics/ocm/labels"
	common2 "github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard"
)

var (
	Names = names.Components
	Verb  = verbs.Add
)

type Command struct {
	utils.BaseCommand

	Force   bool
	Create  bool
	Closure bool

	Handler ctf.FormatHandler
	Format  string

	Version string
	Envs    []string

	Archive string

	Elements []addhdlrs.ElementSource
}

func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx,
		formatoption.New(ctf.GetFormats()...),
		fileoption.New("transport-archive"),
		schemaoption.New(compdesc.DefaultSchemeVersion),
		templateroption.New(""),
		dryrunoption.New("evaluate and print component specifications", true),
		lookupoption.New(),
		rscbyvalueoption.New()),
	}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "[<options>] [--version <version>] [<ctf archive>] {<component-constructor.yaml>}",
		Args:  cobra.MinimumNArgs(1),
		Short: "add component version(s) to a (new) transport archive",
		Example: `
<pre>
$ ocm add componentversions --file ctf --version 1.0 component-constructor.yaml
</pre>

and a file <code>component-constructor.yaml</code>:

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
Add component versions specified by a constructor file to a Common Transport
Archive. The archive might be either a directory prepared to host component version
content or a tar/tgz file (see option --type).

If option <code>--create</code> is given, the archive is created first. An
additional option <code>--force</code> will recreate an empty archive if it
already exists.

If option <code>--complete</code> is given all component versions referenced by
the added one, will be added, also. Therefore, the <code>--lookup</code> is required
to specify an OCM repository to lookup the missing component versions. If 
additionally the <code>-V</code> is given, the resources of those additional
components will be added by value.

The source, resource and reference list can be composed according to the commands
<CMD>ocm add sources</CMD>, <CMD>ocm add resources</CMD>, <CMD>ocm add references</CMD>,
respectively.

The description file might contain:
- a single component as shown in the example
- a list of components under the key <code>components</code>
- a list of yaml documents with a single component or component list

The optional field <code>meta.configuredSchemaVersion</code> for a component
entry can be used to specify a dedicated serialization format to use for the
component descriptor. If given it overrides the <code>--schema</code> option
of the command. By default, v2 is used.

Various elements support to add arbirary information by using labels
(see <CMD>ocm ocm-labels</CMD>).
`,
	}

	cmd.AddCommand(topicocmlabels.New(o.Context))
	return cmd
}

func (o *Command) AddFlags(fs *pflag.FlagSet) {
	o.BaseCommand.AddFlags(fs)
	fs.BoolVarP(&o.Force, "force", "f", false, "remove existing content")
	fs.BoolVarP(&o.Create, "create", "c", false, "(re)create archive")
	fs.BoolVarP(&o.Closure, "complete", "C", false, "include all referenced component version")
	fs.StringArrayVarP(&o.Envs, "settings", "s", nil, "settings file with variable settings (yaml)")
	fs.StringVarP(&o.Version, "version", "v", "", "default version for components")
}

func (o *Command) Complete(args []string) error {
	if o.Closure && !lookupoption.From(o).IsGiven() && o.OCMContext().GetResolver() == nil {
		return fmt.Errorf("lookup option required for option --complete")
	}
	o.Archive, args = fileoption.From(o).GetPath(args, o.Context.FileSystem())

	t := templateroption.From(o)
	err := t.ParseSettings(o.Context.FileSystem(), o.Envs...)
	if err != nil {
		return err
	}

	paths := t.FilterSettings(args...)
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
	session := ocm.NewSession(nil)
	defer session.Close()

	err := o.OptionSet.ProcessOnOptions(common.CompleteOptionsWithSession(o.Context, session))
	if err != nil {
		return err
	}

	printer := common2.NewPrinter(o.Context.StdOut())
	fs := o.Context.FileSystem()
	h := comp.New(o.Version, schemaoption.From(o).Schema)
	elems, ictx, err := addhdlrs.ProcessDescriptions(o.Context, printer, templateroption.From(o).Options, h, o.Elements)
	if err != nil {
		return err
	}

	dr := dryrunoption.From(o)
	if dr.DryRun {
		return addhdlrs.PrintElements(printer, elems, dr.Outfile, o.Context.FileSystem())
	}

	mode := formatoption.From(o).Mode()
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
			o.Create = true
		}
	}

	openmode := accessobj.ACC_WRITABLE
	if o.Create {
		openmode |= accessobj.ACC_CREATE
	}
	repo, err := ctf.Open(o.Context.OCMContext(), openmode, fp, mode, o.Handler, fs)
	if err != nil {
		return err
	}

	thdlr, err := standard.New(standard.KeepGlobalAccess(), standard.Recursive(), rscbyvalueoption.From(o))
	if err != nil {
		return err
	}

	if err == nil {
		err = comp.ProcessComponents(o.Context, ictx, repo, general.Conditional(o.Closure, lookupoption.From(o).Resolver, nil), thdlr, h, elems)
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
