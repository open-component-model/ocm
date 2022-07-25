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

	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/pkg/out"

	"github.com/open-component-model/ocm/cmds/ocm/clictx"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/common/handlers/artefacthdlr"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/common/options/repooption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/transfer"
)

var (
	Names = names.Artefacts
	Verb  = verbs.Transfer
)

type Command struct {
	utils.BaseCommand

	Refs   []string
	Target string
}

func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx, repooption.New())}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] {<artefact-reference>}",
		Args:  cobra.MinimumNArgs(1),
		Short: "transfer OCI artefacts",
		Long: `
Transfer OCI artefacts from one registry to another one
`,
		Example: `
$ ocm oci transfer ghcr.io/mandelsoft/kubelink gcr.io
`,
	}
}

func (o *Command) Complete(args []string) error {
	if len(args) == 0 && repooption.From(o).Spec == "" {
		return fmt.Errorf("a repository or at least one argument that defines the reference is needed")
	}
	o.Target = args[len(args)-1]
	o.Refs = args[:len(args)-1]
	return nil
}

func (o *Command) Run() error {
	session := oci.NewSession(nil)
	defer session.Close()
	err := o.ProcessOnOptions(common.CompleteOptionsWithContext(o.Context, session))
	if err != nil {
		return err
	}
	a, err := NewAction(o.Context, session, o.Target)
	if err != nil {
		return err
	}

	handler := artefacthdlr.NewTypeHandler(o.Context.OCI(), a.Session, repooption.From(o).Repository)

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
	repo, err := session.DetermineRepositoryBySpec(ctx.OCIContext(), &ref.UniformRepositorySpec)
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
	defer ns.Close()
	tgt := a.Ref
	tgt.Repository = ns.GetNamespace()
	if tag != "" {
		tgt.Tag = &tag
	}
	out.Outf(a.Context, "copying %s to %s...\n", &src.Spec, &tgt)
	err = transfer.TransferArtefact(src.Artefact, ns, tag)
	if err == nil {
		a.copied++
	}
	return err
}

func (a *action) Out() error {
	out.Outf(a.Context, "copied %d from %d artefact(s)\n", a.copied, a.count)
	return nil
}

func (a *action) Target(obj *artefacthdlr.Object) (string, string) {
	if a.Ref.IsRegistry() {
		if obj.Spec.Tag != nil {
			return obj.Spec.Repository, *obj.Spec.Tag
		}
		return obj.Spec.Repository, ""
	}
	if a.Ref.IsVersion() {
		return a.Ref.Repository, *a.Ref.Tag
	}
	if obj.Spec.Tag != nil {
		return a.Ref.Repository, *obj.Spec.Tag
	}
	return a.Ref.Repository, ""
}
