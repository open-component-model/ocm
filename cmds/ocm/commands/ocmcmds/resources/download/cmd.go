// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package download

import (
	"fmt"
	"io"
	"path"
	"runtime"
	"strings"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/closureoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/destoption"
	ocmcommon "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/handlers/elemhdlr"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/downloaderoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/versionconstraintsoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/resources/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	common2 "github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/consts"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/out"
)

var (
	Names = names.Resources
	Verb  = verbs.Download
)

type Command struct {
	utils.BaseCommand

	Executable    bool
	ResourceTypes []string

	Comp string
	Ids  []v1.Identity
}

// NewCommand creates a new resources command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	f := func(opts *output.Options) output.Output {
		return &action{downloaders: download.For(ctx.OCMContext()), opts: opts}
	}
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx,
		versionconstraintsoption.New(),
		repooption.New(),
		downloaderoption.New(),
		output.OutputOptions(output.NewOutputs(f), NewOptions(), closureoption.New("component reference"), lookupoption.New(), destoption.New()),
	)}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>]  <component> {<name> { <key>=<value> }}",
		Args:  cobra.MinimumNArgs(1),
		Short: "download resources of a component version",
		Long: `
Download resources of a component version. Resources are specified
by identities. An identity consists of 
a name argument followed by optional <code>&lt;key>=&lt;value></code>
arguments.

The option <code>-O</code> is used to declare the output destination.
For a single resource to download, this is the file written for the
resource blob. If multiple resources are selected, a directory structure
is written into the given directory for every involved component version
as follows:

<center>
    <pre>&lt;component>/&lt;version>{/&lt;nested component>/&lt;version>}</pre>
</center>

The resource files are named according to the resource identity in the
component descriptor. If this identity is just the resource name, this name
is ised. If additional identity attributes are required, this name is
append by a comma separated list of <code>&lt;name>=&lt;>value></code> pairs
separated by a "-" from the plain name. This attribute list is alphabetical
order:

<center>
    <pre>&lt;resource name>[-[&lt;name>=&lt;>value>]{,&lt;name>=&lt;>value>}]</pre>
</center>

`,
	}
}

func (o *Command) AddFlags(fs *pflag.FlagSet) {
	o.BaseCommand.AddFlags(fs)
	fs.BoolVarP(&o.Executable, "executable", "x", false, "download executable for local platform")
	fs.StringArrayVarP(&o.ResourceTypes, "type", "t", nil, "resource type filter")
}

func (o *Command) Complete(args []string) error {
	var err error
	o.Comp = args[0]
	o.Ids, err = ocmcommon.MapArgsToIdentities(args[1:]...)
	if err == nil && o.Executable {
		if len(o.ResourceTypes) == 0 {
			o.ResourceTypes = []string{resourcetypes.EXECUTABLE}
		}
		if len(o.Ids) == 0 {
			o.Ids = []v1.Identity{
				{},
			}
		}
		for _, id := range o.Ids {
			id[consts.ExecutableOperatingSystem] = runtime.GOOS
			id[consts.ExecutableArchitecture] = runtime.GOARCH
		}
	}
	return err
}

func (o *Command) handlerOptions() []elemhdlr.Option {
	hopts := common.OptionsFor(o)
	if len(o.ResourceTypes) > 0 {
		hopts = append(hopts, common.WithTypes(o.ResourceTypes))
	}
	return hopts
}

func (o *Command) Run() error {
	session := ocm.NewSession(nil)
	defer session.Close()

	err := o.ProcessOnOptions(ocmcommon.CompleteOptionsWithSession(o, session))
	if err != nil {
		return err
	}

	d := downloaderoption.From(o)
	err = d.Register(o)
	if err != nil {
		return err
	}

	opts := output.From(o)
	if d.HasRegistrations() || o.Executable {
		From(opts).UseHandlers = true
	}

	hdlr, err := common.NewTypeHandler(o.Context.OCM(), opts, repooption.From(o).Repository, session, []string{o.Comp}, o.handlerOptions()...)
	if err != nil {
		return err
	}
	specs, err := utils.ElemSpecs(o.Ids)
	if err != nil {
		return err
	}

	return utils.HandleOutputs(opts, hdlr, specs...)
}

////////////////////////////////////////////////////////////////////////////////

type action struct {
	downloaders download.Registry
	data        elemhdlr.Objects
	opts        *output.Options
}

func (d *action) Add(e interface{}) error {
	d.data = append(d.data, e.(*elemhdlr.Object))
	return nil
}

func (d *action) Close() error {
	return nil
}

func (d *action) Out() error {
	list := errors.ErrListf("downloading resources")
	dest := destoption.From(d.opts)
	if len(d.data) == 1 {
		if dest.Destination == "" {
			_, _ = common.Elem(d.data[0]).Labels.GetValue("downloadName", &dest.Destination)
		}
		return d.Save(d.data[0], dest.Destination)
	} else {
		if dest.Destination == "-" {
			return fmt.Errorf("standard output supported for single resource only.")
		}
		for _, e := range d.data {
			f := dest.Destination
			if f == "" {
				f = "."
			}
			for _, p := range e.History {
				f = path.Join(f, p.GetName(), p.GetVersion())
			}
			r := common.Elem(e)
			n := ""
			if ok, err := r.Labels.GetValue("downloadName", &n); !ok || err != nil {
				n = r.Name
			}
			f = path.Join(f, n)
			id := r.GetIdentity(e.Version.GetDescriptor().Resources)
			delete(id, v1.SystemIdentityName)
			if len(id) > 0 {
				f += "-" + strings.ReplaceAll(id.String(), "\"", "")
			}
			err := d.Save(e, f)
			if err != nil {
				list.Add(err)
				out.Outf(d.opts.Context, "%s failed: %s\n", f, err)
			}
		}
	}
	return list.Result()
}

func (d *action) Save(o *elemhdlr.Object, f string) error {
	printer := common2.NewPrinter(d.opts.Context.StdOut())
	dest := destoption.From(d.opts)
	local := From(d.opts)
	pathIn := true
	r := common.Elem(o)
	if f == "" {
		f = r.GetName()
		pathIn = false
	}
	var tmp vfs.File
	var err error
	if f == "-" {
		tmp, err = vfs.TempFile(dest.PathFilesystem, "", "download-*")
		if err != nil {
			return err
		}
		f = tmp.Name()
		tmp.Close()
		printer = common2.NewPrinter(nil)
		defer dest.PathFilesystem.Remove(f)
	}
	id := r.GetIdentity(o.Version.GetDescriptor().Resources)
	racc, err := o.Version.GetResource(id)
	if err != nil {
		return err
	}
	dir := path.Dir(f)
	err = dest.PathFilesystem.MkdirAll(dir, 0o770)
	if err != nil {
		return err
	}
	var ok bool
	var eff string
	if local.UseHandlers {
		ok, eff, err = d.downloaders.Download(printer, racc, f, dest.PathFilesystem)
	} else {
		ok, eff, err = d.downloaders.DownloadAsBlob(printer, racc, f, dest.PathFilesystem)
	}
	if err != nil {
		return err
	}
	if !ok {
		return errors.Newf("no downloader configured for type %q", racc.Meta().GetType())
	}
	if tmp != nil {
		if eff != f {
			defer dest.PathFilesystem.Remove(eff)
		}
		file, err := dest.PathFilesystem.Open(eff)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(d.opts.Context.StdOut(), file)
		if err != nil {
			return err
		}
	} else if eff != f && pathIn {
		out.Outf(d.opts.Context, "output path %q changed to %q by downloader", f, eff)
	}
	return nil
}
