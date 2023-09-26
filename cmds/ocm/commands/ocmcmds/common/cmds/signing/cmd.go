// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package signing

import (
	"fmt"

	"github.com/spf13/cobra"

	ocmcommon "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/handlers/comphdlr"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/signoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/versionconstraintsoption"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
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
	return utils.SetupCommand(&SignatureCommand{spec: spec, BaseCommand: utils.NewBaseCommand(ctx, versionconstraintsoption.New(), repooption.New(), signoption.New(sign), lookupoption.New())}, names...)
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

func (o *SignatureCommand) Run() (rerr error) {
	session := ocm.NewSession(nil)
	defer errors.PropagateError(&rerr, func() error {
		return session.Close()
	})

	err := o.ProcessOnOptions(ocmcommon.CompleteOptionsWithSession(o, session))
	if err != nil {
		return err
	}
	sign := signoption.From(o)
	repo := repooption.From(o).Repository
	lookup := lookupoption.From(o)
	handler := comphdlr.NewTypeHandler(o.Context.OCM(), session, repo, comphdlr.OptionsFor(o))
	sopts := signing.NewOptions(sign, signing.Resolver(repo, lookup.Resolver))
	err = sopts.Complete(o.Context.OCMContext())
	if err != nil {
		return err
	}
	return utils.HandleOutput(NewAction(o.spec.terms, o.Context.OCMContext(), common.NewPrinter(o.Context.StdOut()), sopts), handler, utils.StringElemSpecs(o.Refs...)...)
}

/////////////////////////////////////////////////////////////////////////////

type Action interface {
	output.Output
	Digest(o *comphdlr.Object) (*metav1.DigestSpec, *compdesc.ComponentDescriptor, error)
}

type action struct {
	desc         []string
	printer      common.Printer
	state        signing.WalkingState
	baseresolver ocm.ComponentVersionResolver
	sopts        *signing.Options
	errlist      *errors.ErrorList
}

var _ output.Output = (*action)(nil)

func NewAction(desc []string, ctx ocm.Context, p common.Printer, sopts *signing.Options) Action {
	return &action{
		desc:         desc,
		printer:      p,
		state:        signing.NewWalkingState(ctx.LoggingContext().WithContext(signing.REALM)),
		baseresolver: sopts.Resolver,
		sopts:        sopts,
		errlist:      errors.ErrListf(desc[1]),
	}
}

func (a *action) Digest(o *comphdlr.Object) (*metav1.DigestSpec, *compdesc.ComponentDescriptor, error) {
	sopts := *a.sopts
	sopts.Resolver = ocm.NewCompoundResolver(o.Repository, a.sopts.Resolver)
	d, err := signing.Apply(a.printer, &a.state, o.ComponentVersion, &sopts)
	var cd *compdesc.ComponentDescriptor
	nv := common.VersionedElementKey(o.ComponentVersion)
	vi := a.state.Get(nv)
	if vi != nil {
		cd = vi.GetContext(nv).Descriptor
	}
	return d, cd, err
}

func (a *action) Add(e interface{}) error {
	o, ok := e.(*comphdlr.Object)
	if !ok {
		return fmt.Errorf("failed to assert %T to *comphdlr.Object", e)
	}
	cv := o.ComponentVersion
	d, _, err := a.Digest(o)
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
