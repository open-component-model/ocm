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

package signing

import (
	"fmt"

	"github.com/spf13/cobra"

	ocmcommon "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/handlers/comphdlr"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/signoption"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/signingattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/signing"
	"github.com/open-component-model/ocm/pkg/errors"
)

type SignatureCommand struct {
	utils.BaseCommand
	Refs []string
	spec *spec
}

type spec struct {
	op      string
	sign    bool
	example string
	terms   []string
}

func newOperation(op string, sign bool, terms []string, example string) *spec {
	return &spec{
		op:      op,
		sign:    sign,
		example: example,
		terms:   terms,
	}
}

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, op string, sign bool, terms []string, example string, names ...string) *cobra.Command {
	spec := newOperation(op, sign, terms, example)
	return utils.SetupCommand(&SignatureCommand{spec: spec, BaseCommand: utils.NewBaseCommand(ctx, repooption.New(), signoption.New(sign))}, names...)
}

func (o *SignatureCommand) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] {<component-reference>}",
		Short: o.spec.op + " component version",
		Long: `
` + o.spec.op + ` specified component versions. 
`,
		Example: o.spec.example,
	}
}

func (o *SignatureCommand) Complete(args []string) error {
	o.Refs = args
	if len(args) == 0 && repooption.From(o).Spec == "" {
		return fmt.Errorf("a repository or at least one argument that defines the reference is needed")
	}
	return nil
}

func (o *SignatureCommand) Run() error {
	session := ocm.NewSession(nil)
	defer session.Close()

	err := o.ProcessOnOptions(ocmcommon.CompleteOptionsWithSession(o, session))
	if err != nil {
		return err
	}
	sign := signoption.From(o)
	repo := repooption.From(o).Repository
	handler := comphdlr.NewTypeHandler(o.Context.OCM(), session, repo)
	sopts := signing.NewOptions(sign, signing.Resolver(repo))
	err = sopts.Complete(signingattr.Get(o.Context.OCMContext()))
	if err != nil {
		return err
	}
	return utils.HandleOutput(NewAction(o.spec.terms, o, sopts), handler, utils.StringElemSpecs(o.Refs...)...)
}

/////////////////////////////////////////////////////////////////////////////

type action struct {
	desc         []string
	cmd          *SignatureCommand
	printer      common.Printer
	state        common.WalkingState
	baseresolver ocm.ComponentVersionResolver
	sopts        *signing.Options
	errlist      *errors.ErrorList
}

var _ output.Output = (*action)(nil)

func NewAction(desc []string, cmd *SignatureCommand, sopts *signing.Options) output.Output {
	return &action{
		desc:         desc,
		cmd:          cmd,
		printer:      common.NewPrinter(cmd.Context.StdOut()),
		state:        common.NewWalkingState(),
		baseresolver: sopts.Resolver,
		sopts:        sopts,
		errlist:      errors.ErrListf(desc[1]),
	}
}

func (a *action) Add(e interface{}) error {
	o := e.(*comphdlr.Object)
	cv := o.ComponentVersion
	sopts := *a.sopts
	sopts.Resolver = ocm.NewCompoundResolver(o.Repository, a.sopts.Resolver)
	d, err := signing.Apply(a.printer, &a.state, cv, &sopts)
	a.errlist.Add(err)
	if err == nil {
		a.printer.Printf("successfully %s %s:%s (digest %s:%s)\n", a.desc[0], cv.GetName(), cv.GetVersion(), d.HashAlgorithm, d.Value)
	} else {
		a.printer.Printf("failed %s %s:%s: %s\n", a.desc[1], cv.GetName(), cv.GetVersion(), err)
	}
	return nil
}

func (a *action) Close() error {
	return nil
}

func (a *action) Out() error {
	if a.errlist.Len() > 0 {
		a.printer.Printf("finished with %d error(s)\n", a.errlist.Len())
	}
	return a.errlist.Result()
}
