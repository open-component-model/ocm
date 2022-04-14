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
	"io"
	"path"
	"strings"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/cmds/ocm/clictx"
	"github.com/open-component-model/ocm/cmds/ocm/commands"
	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/closureoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/destoption"
	ocmcommon "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/handlers/elemhdlr"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/resources/common"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output/out"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/ocm"
	metav1 "github.com/open-component-model/ocm/pkg/ocm/compdesc/meta/v1"
	"github.com/spf13/cobra"
)

var (
	Names = names.Resources
	Verb  = commands.Download
)

type Command struct {
	utils.BaseCommand

	Comp string
	Ids  []metav1.Identity
}

// NewCommand creates a new resources command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx, &repooption.Option{}, output.OutputOptions(outputs, closureoption.New("component reference"), &lookupoption.Option{}, &destoption.Option{}))}, names...)
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
<code>&lt;component>/&lt;version>{/&lt;nested component>/&lt;version>}</code>
</center>

The resource files are named according to the resource identity in the
component descriptor. If this identity is just the resource name, this name
is ised. If additional identity attributes are required, this name is
append by a comma separated list of <code>&lt;name>=&lt>value></code> pairs
separated by a "-" from the plain name. This attribute list is alphabetical
order:

<center>
<code>&lt;resource name>[-[&lt;name>=&lt>value>]{,&lt;name>=&lt>value>}]</code>
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

	err := o.ProcessOnOptions(ocmcommon.CompleteOptionsWithContext(o, session))
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

var outputs = output.NewOutputs(get_download)

func get_download(opts *output.Options) output.Output {
	return &download{opts: opts}
}

type download struct {
	data elemhdlr.Objects
	opts *output.Options
}

func (d *download) Add(e interface{}) error {
	d.data = append(d.data, e.(*elemhdlr.Object))
	return nil
}

func (d *download) Close() error {
	return nil
}

func (d *download) Out() error {
	list := errors.ErrListf("downloading resources")
	dest := destoption.From(d.opts)
	if len(d.data) == 1 {
		return d.Save(d.data[0], dest.Destination)
	} else {
		for _, e := range d.data {
			f := dest.Destination
			for _, p := range e.History {
				f = path.Join(f, p.GetName(), p.GetVersion())
			}
			r := common.Elem(e)
			f = path.Join(f, r.Name)
			id := r.GetIdentity(e.Version.GetDescriptor().Resources)
			delete(id, metav1.SystemIdentityName)
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

func (d *download) Save(o *elemhdlr.Object, f string) error {
	dest := destoption.From(d.opts)
	r := common.Elem(o)
	id := r.GetIdentity(o.Version.GetDescriptor().Resources)
	acc, err := o.Version.GetResource(id)
	if err != nil {
		return err
	}
	dir := path.Dir(f)
	err = dest.PathFilesystem.MkdirAll(dir, 0770)
	if err != nil {
		return err
	}
	rd, err := acc.Reader()
	if err != nil {
		return err
	}
	defer rd.Close()
	file, err := dest.PathFilesystem.OpenFile(f, vfs.O_TRUNC|vfs.O_CREATE|vfs.O_WRONLY, 0660)
	if err != nil {
		return err
	}
	defer file.Close()
	n, err := io.Copy(file, rd)
	if err == nil {
		out.Outf(d.opts.Context, "%s: %d byte(s) written\n", f, n)
	}
	return err
}
