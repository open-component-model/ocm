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

	ocmcommon "github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/common/options/repooption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/out"
	utils2 "github.com/open-component-model/ocm/pkg/utils"
)

var (
	Names = names.Tags
	Verb  = verbs.Show
)

type Command struct {
	utils.BaseCommand
	Latest   bool
	Semantic bool
	Semver   bool

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
		Short: "show dedicated tags of OCI artefacts",
		Long: `
Match tags of an artefact against some patterns.
`,
		Example: `
$ oci show tags ghcr.io/mandelsoft/kubelink
`,
	}
}

func (o *Command) AddFlags(fs *pflag.FlagSet) {
	o.BaseCommand.AddFlags(fs)
	fs.BoolVarP(&o.Latest, "latest", "l", false, "show only latest tags")
	fs.BoolVarP(&o.Semver, "semver", "s", false, "show only semver compliant tags")
	fs.BoolVarP(&o.Semantic, "semantic", "o", false, "show semantic tags")
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

	if o.Semantic {
		o.Semver = true
	}
	return nil
}

func (o *Command) Run() error {
	session := oci.NewSession(nil)
	defer session.Close()

	err := o.ProcessOnOptions(ocmcommon.CompleteOptionsWithContext(o, session))
	if err != nil {
		return err
	}

	versions := Versions{}
	tags := utils2.StringSlice{}
	repo := repooption.From(o)

	var art oci.ArtefactAccess
	var ns oci.NamespaceAccess

	// determine version source
	if repo.Repository != nil {
		cr, err := oci.ParseArt(o.Ref)
		if err != nil {
			return err
		}
		ns, err = session.LookupNamespace(repo.Repository, cr.Repository)
		if err != nil {
			return err
		}
		if cr.IsVersion() {
			art, err = session.GetArtefact(ns, cr.Reference())
			if err != nil {
				return err
			}
		}
	} else {
		r, err := session.EvaluateRef(o.Context.OCIContext(), o.Ref)
		if err != nil {
			return err
		}
		if r.Namespace == nil {
			return errors.Newf("no namespace specified")
		}
		ns = r.Namespace
		art = r.Artefact
	}

	list, err := ns.ListTags()
	if err != nil {
		return err
	}
	tags = utils2.StringSlice(list)
	// determine version base set
	if art != nil {
		dig := art.Digest()
		for i := 0; i < len(tags); i++ {
			a, err := ns.GetArtefact(tags[i])
			if err != nil {
				return err
			}
			if a.Digest() != dig {
				tags.Delete(i)
				i--
			} else {
				v, err := semver.NewVersion(tags[i])
				if err == nil {
					versions = append(versions, v)
				}
			}
		}
	} else {
		tags, err = ns.ListTags()
		if err != nil {
			return err
		}
		for _, vn := range tags {
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
	tags.Sort()
	if len(versions) > 1 && o.Latest {
		versions = versions[len(versions)-1:]
	}

	if o.Semver {
		for _, r := range versions {
			if o.Semantic {
				out.Outf(o, "%s\n", r)
			} else {
				out.Outf(o, "%s\n", r.Original())
			}
		}
	} else {
		for _, r := range tags {
			out.Outf(o, "%s\n", r)
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
