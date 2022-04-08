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

package transfer

import (
	"fmt"

	"github.com/gardener/ocm/cmds/ocm/commands"
	"github.com/gardener/ocm/cmds/ocm/commands/ocicmds/artefacts/common/options/repooption"
	"github.com/gardener/ocm/cmds/ocm/commands/ocicmds/common/handlers/artefacthdlr"
	"github.com/gardener/ocm/cmds/ocm/commands/ocicmds/names"
	"github.com/gardener/ocm/pkg/oci"
	"github.com/gardener/ocm/pkg/oci/transfer"

	"github.com/gardener/ocm/cmds/ocm/clictx"
	"github.com/gardener/ocm/cmds/ocm/pkg/output"
	"github.com/gardener/ocm/cmds/ocm/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	Names = names.Artefacts
	Verb  = commands.Transfer
)

type Command struct {
	utils.BaseCommand

	Output output.Options

	Repository repooption.Option
	Refs       []string
	Target     string
}

func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx)}, names...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] {<artefact-reference>}",
		Args:  cobra.MinimumNArgs(1),
		Short: "transfer OCI artefacts",
		Long: `
Transfer OCI artefacts from one registry to another one

` + o.Repository.Usage() + `

*Example:*
<pre>
$ ocm oci transfer ghcr.io/mandelsoft/kubelink gcr.io
</pre>
`,
	}
}

func (o *Command) AddFlags(fs *pflag.FlagSet) {
	o.Repository.AddFlags(fs)
}

func (o *Command) Complete(args []string) error {
	var err error
	if len(args) == 0 && o.Repository.Spec == "" {
		return fmt.Errorf("a repository or at least one argument that defines the reference is needed")
	}
	o.Target = args[len(args)-1]
	o.Refs = args[:len(args)-1]
	err = o.Repository.Complete(o.Context)
	if err != nil {
		return err
	}
	return o.Output.Complete(o.Context)

}

func (o *Command) Run() error {
	session := oci.NewSession(nil)
	defer session.Close()
	repo, err := o.Repository.GetRepository(o.Context.OCI(), session)
	if err != nil {
		return err
	}
	a, err := NewAction(o.Context, session, o.Target)
	if err != nil {
		return err
	}

	handler := artefacthdlr.NewTypeHandler(o.Context.OCI(), a.Session, repo)

	return utils.HandleOutput(a, handler, utils.StringElemSpecs(o.Refs...)...)
}

/////////////////////////////////////////////////////////////////////////////

type action struct {
	oci.Session

	count    int
	copied   int
	Context  clictx.Context
	Registry oci.Repository
	Ref      oci.RefSpec
}

func NewAction(ctx clictx.Context, session oci.Session, target string) (*action, error) {
	ref, err := oci.ParseRef(target)
	if err != nil {
		return nil, err
	}
	if ref.Digest != nil {
		return nil, fmt.Errorf("copy to target digest not supported")
	}
	repo, err := session.DetermineRepositoryBySpec(ctx.OCIContext(), &ref.UniformRepositorySpec, ctx.OCI().GetAlias)
	if err != nil {
		return nil, err
	}

	return &action{
		Context:  ctx,
		Session:  session,
		Ref:      ref,
		Registry: repo,
	}, nil
}

func (a *action) Add(e interface{}) error {
	if a.count > 0 && !a.Ref.IsRegistry() {
		return fmt.Errorf("cannot copy multiple source artefacts to the same artefact (%s)", &a.Ref)
	}
	a.count++

	src := e.(*artefacthdlr.Object)

	repository, tag := a.Target(src)

	ns, err := a.Registry.LookupNamespace(repository)
	if err != nil {
		return err
	}
	tgt := a.Ref
	tgt.Repository = ns.GetNamespace()
	if tag != "" {
		tgt.Tag = &tag
	}
	fmt.Printf("copying %s to %s...\n", &src.Spec, &tgt)
	err = transfer.TransferArtefact(src.Artefact, ns, tag)
	if err == nil {
		a.copied++
	}
	return err
}

func (a *action) Out() error {
	fmt.Printf("copied %d from %d artefact(s)\n", a.copied, a.count)
	return nil
}

func (a *action) Target(obj *artefacthdlr.Object) (string, string) {
	if a.Ref.IsRegistry() {
		if a.Ref.IsVersion() {
			return a.Ref.Repository, *a.Ref.Tag
		}
		return a.Ref.Repository, ""
	}
	if a.Ref.IsVersion() {
		return a.Ref.Repository, *a.Ref.Tag
	}
	if obj.Spec.Tag != nil {
		return a.Ref.Repository, *obj.Spec.Tag
	}
	return a.Ref.Repository, ""
}
