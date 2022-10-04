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

package download

import (
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/closureoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/destoption"
	ocmcommon "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/handlers/elemhdlr"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/resources/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/out"
)

var (
	Names = names.Resources
	Verb  = verbs.Download
)

type Command struct {
	utils.BaseCommand

	Comp string
	Ids  []v1.Identity
}

// NewCommand creates a new resources command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	f := func(opts *output.Options) output.Output {
		return &action{downloaders: download.For(ctx), opts: opts}
	}
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx, repooption.New(), output.OutputOptions(output.NewOutputs(f), NewOptions(), closureoption.New("component reference"), lookupoption.New(), destoption.New()))}, utils.Names(Names, names...)...)
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
append by a comma separated list of <code>&lt;name>=&lt>value></code> pairs
separated by a "-" from the plain name. This attribute list is alphabetical
order:

<center>
    <pre>&lt;resource name>[-[&lt;name>=&lt>value>]{,&lt;name>=&lt>value>}]</pre>
</center>

`,
	}
}

func (o *Command) Complete(args []string) error {
	var err error
	o.Comp = args[0]
	o.Ids, err = ocmcommon.MapArgsToIdentities(args[1:]...)
	return err
}

func (o *Command) Run() error {
	session := ocm.NewSession(nil)
	defer session.Close()

	err := o.ProcessOnOptions(ocmcommon.CompleteOptionsWithSession(o, session))
	if err != nil {
		return err
	}

	opts := output.From(o)
	hdlr, err := common.NewTypeHandler(o.Context.OCM(), opts, repooption.From(o).Repository, session, []string{o.Comp})
	if err != nil {
		return err
	}
	return utils.HandleOutputs(opts, hdlr, utils.ElemSpecs(o.Ids)...)
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
		return d.Save(d.data[0], dest.Destination)
	} else {
		if dest.Destination == "-" {
			return fmt.Errorf("standard output suported for singlle resource only.")
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
			f = path.Join(f, r.Name)
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
		ok, eff, err = d.downloaders.Download(d.opts.Context, racc, f, dest.PathFilesystem)
	} else {
		ok, eff, err = d.downloaders.DownloadAsBlob(d.opts.Context, racc, f, dest.PathFilesystem)
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
