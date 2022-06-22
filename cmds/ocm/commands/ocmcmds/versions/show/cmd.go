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

package show

import (
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/out"

	"github.com/open-component-model/ocm/cmds/ocm/clictx"
	ocmcommon "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
)

var (
	Names = names.Versions
	Verb  = verbs.Show
)

type Command struct {
	utils.BaseCommand
	Latest   bool
	Semantic bool

	Ref         string
	Constraints []*semver.Constraints
}

// NewCommand creates a new ocm command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx,
		repooption.New(),
	)}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] <component> {<version pattern>}",
		Args:  cobra.MinimumNArgs(1),
		Short: "show dedicated versions (semver compliant)",
		Long: `
Match versions of a component against some patterns.
`,
		Example: `
$ ocm show versions ghcr.io/mandelsoft/cnudie//github.com/mandelsoft/playground
`,
	}
}

func (o *Command) AddFlags(fs *pflag.FlagSet) {
	o.BaseCommand.AddFlags(fs)
	fs.BoolVarP(&o.Latest, "latest", "l", false, "show only latest version")
	fs.BoolVarP(&o.Semantic, "semantic", "s", false, "show semantic version")
}

func (o *Command) Complete(args []string) error {
	o.Ref = args[0]

	for _, v := range args[1:] {
		c, err := semver.NewConstraint(v)
		if err != nil {
			return err
		}
		o.Constraints = append(o.Constraints, c)
	}
	return nil
}

func (o *Command) Run() error {
	session := ocm.NewSession(nil)
	defer session.Close()

	err := o.ProcessOnOptions(ocmcommon.CompleteOptionsWithContext(o, session))
	if err != nil {
		return err
	}

	versions := Versions{}
	repo := repooption.From(o)

	var cv ocm.ComponentVersionAccess
	var comp ocm.ComponentAccess

	// determine version source
	if repo.Repository != nil {
		cr, err := ocm.ParseComp(o.Ref)
		if err != nil {
			return err
		}
		comp, err = session.LookupComponent(repo.Repository, cr.Component)
		if err != nil {
			return err
		}
		if cr.IsVersion() {
			cv, err = session.GetComponentVersion(comp, *cr.Version)
			if err != nil {
				return err
			}
		}
	} else {
		r, err := session.EvaluateVersionRef(o.Context.OCMContext(), o.Ref)
		if err != nil {
			return err
		}
		if r.Component == nil {
			return errors.Newf("no component specified")
		}
		comp = r.Component
		cv = r.Version
	}

	// determine version base set
	if cv != nil {
		v, err := semver.NewVersion(cv.GetVersion())
		if err != nil {
			return err
		}
		versions = append(versions, v)
	} else {
		vers, err := comp.ListVersions()
		if err != nil {
			return err
		}
		for _, vn := range vers {
			v, err := semver.NewVersion(vn)
			if err == nil {
				versions = append(versions, v)
			}
		}
	}

	// Filter by patterns
	for i := 0; i < len(versions); i++ {
		found := len(o.Constraints) == 0
		for _, c := range o.Constraints {
			if c.Check(versions[i]) {
				found = true
			}
		}
		if !found {
			versions = append(versions[:i], versions[i+1:]...)
			i--
		}
	}

	sort.Sort(versions)
	if len(versions) > 1 && o.Latest {
		versions = versions[len(versions)-1:]
	}

	for _, r := range versions {
		if o.Semantic {
			out.Outf(o, "%s\n", r)
		} else {
			out.Outf(o, "%s\n", r.Original())
		}
	}
	return nil
}

type Versions = semver.Collection

/*
var _ sort.Interface = (Versions)(nil)

func (v Versions) Len() int {
	return len(v)
}

func (v Versions) Less(i, j int) bool {
	return v[i].Compare(v[j])<0
}

func (v Versions) Swap(i, j int) {
	v[i],v[j]=v[j],v[i]
}

*/
