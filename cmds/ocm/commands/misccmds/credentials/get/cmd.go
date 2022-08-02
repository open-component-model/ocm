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

package get

import (
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/errors"

	"github.com/open-component-model/ocm/cmds/ocm/commands/misccmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"

	"github.com/open-component-model/ocm/cmds/ocm/clictx"
)

var (
	Names = names.Credentials
	Verb  = verbs.Get
)

type Command struct {
	utils.BaseCommand

	Consumer credentials.ConsumerIdentity
	Matcher  credentials.IdentityMatcher

	Type string
}

var _ utils.OCMCommand = (*Command)(nil)

// NewCommand creates a new artefact command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx)}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {

	return &cobra.Command{
		Use:   "{<consumer property>=<value>}",
		Short: "Get credentials for a dedicated consumer spec",
		Long: `
Try to resolve a given consumer specification against the configured credential
settings and show the found credential attributes.

For the following usage contexts with matchers and standard identity matchers exist:
` + utils.FormatListElements("", o.CredentialsContext().ConsumerIdentityMatchers().List()) +
			`
The used matcher is derived from the consumer attribute <code>type</code>.
For all other consumer types a matcher matching all attributes will be used.
The usage of a dedicated matcher can be enforced by the option <code>--matcher</code>.
`,
	}
}

func (o *Command) AddFlags(set *pflag.FlagSet) {
	set.StringVarP(&o.Type, "matcher", "m", "", "matcher type override")
}

func (o *Command) Complete(args []string) error {
	if o.Type != "" {
		m, _ := o.CredentialsContext().ConsumerIdentityMatchers().Get(o.Type)
		if m == nil {
			return errors.ErrUnknown("identity matcher", o.Type)
		}
		o.Matcher = m
	}
	o.Consumer = credentials.ConsumerIdentity{}
	for _, s := range args {
		i := strings.Index(s, "=")
		if i < 0 {
			return errors.ErrInvalid("consumer setting", s)
		}
		name := s[:i]
		value := s[i+1:]
		if len(name) == 0 {
			return errors.ErrInvalid("credential setting", s)
		}
		o.Consumer[name] = value
	}
	if t, ok := o.Consumer[credentials.CONSUMER_ATTR_TYPE]; ok {
		m, _ := o.CredentialsContext().ConsumerIdentityMatchers().Get(t)
		if m != nil {
			o.Matcher = m
		}
	}
	if o.Matcher == nil {
		o.Matcher = credentials.PartialMatch
	}
	return nil
}

func (o *Command) Run() error {

	src, err := o.CredentialsContext().GetCredentialsForConsumer(o.Consumer, o.Matcher)
	if err != nil {
		return err
	}
	creds, err := src.Credentials(o.CredentialsContext())
	if err != nil {
		return err
	}

	var list [][]string
	for k, v := range creds.Properties() {
		list = append(list, []string{k, v})
	}
	sort.Slice(list, func(i, j int) bool { return strings.Compare(list[i][0], list[j][0]) < 0 })
	output.FormatTable(o, "", append([][]string{[]string{"ATTRIBUTE", "VALUE"}}, list...))
	return nil
}
