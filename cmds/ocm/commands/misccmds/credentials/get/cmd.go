// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package get

import (
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/commands/misccmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/errors"
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

// NewCommand creates a new artifact command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx)}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	var standard credentials.IdentityMatcherInfos
	var context credentials.IdentityMatcherInfos
	for _, e := range o.CredentialsContext().ConsumerIdentityMatchers().List() {
		if e.Type != "" {
			r, _ := utf8.DecodeRuneInString(e.Type)
			if unicode.IsLower(r) {
				standard = append(standard, e)
			} else {
				context = append(context, e)
			}
		} else {
			context = append(context, e)
		}
	}

	return &cobra.Command{
		Use:   "{<consumer property>=<value>}",
		Short: "Get credentials for a dedicated consumer spec",
		Long: `
Try to resolve a given consumer specification against the configured credential
settings and show the found credential attributes.

Matchers exist for the following usage contexts or consumer types:
` + utils.FormatListElements("", context) +
			`
The following standard identity matchers are supported:
` + utils.FormatListElements("partial", standard) +
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
		m := o.CredentialsContext().ConsumerIdentityMatchers().Get(o.Type)
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
	if t, ok := o.Consumer[credentials.ID_TYPE]; ok {
		m := o.CredentialsContext().ConsumerIdentityMatchers().Get(t)
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
	creds, err := credentials.RequiredCredentialsForConsumer(o.CredentialsContext(), o.Consumer, o.Matcher)
	if err != nil {
		return err
	}

	var list [][]string
	for k, v := range creds.Properties() {
		list = append(list, []string{k, v})
	}
	sort.Slice(list, func(i, j int) bool { return strings.Compare(list[i][0], list[j][0]) < 0 })
	output.FormatTable(o, "", append([][]string{{"ATTRIBUTE", "VALUE"}}, list...))
	return nil
}
