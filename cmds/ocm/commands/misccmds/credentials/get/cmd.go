package get

import (
	"sort"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/utils/listformat"
	"ocm.software/ocm/api/utils/out"
	"ocm.software/ocm/cmds/ocm/commands/misccmds/names"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/output"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var (
	Names = names.Credentials
	Verb  = verbs.Get
)

type Command struct {
	utils.BaseCommand

	Consumer credentials.ConsumerIdentity
	Matcher  credentials.IdentityMatcher

	Type   string
	Sloppy bool
}

var _ utils.OCMCommand = (*Command)(nil)

// NewCommand creates a new artifact command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx)}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	var standard credentials.IdentityMatcherInfos
	var consumer credentials.IdentityMatcherInfos
	for _, e := range o.CredentialsContext().ConsumerIdentityMatchers().List() {
		if e.IsConsumerType() {
			consumer = append(consumer, e)
		} else {
			standard = append(standard, e)
		}
	}

	return &cobra.Command{
		Use:   "{<consumer property>=<value>}",
		Short: "Get credentials for a dedicated consumer spec",
		Long: `
Try to resolve a given consumer specification against the configured credential
settings and show the found credential attributes.

Matchers exist for the following usage contexts or consumer types:
` + listformat.FormatListElements("", consumer) +
			`
The following standard identity matchers are supported:
` + listformat.FormatListElements("partial", standard) +
			`
The used matcher is derived from the consumer attribute <code>type</code>.
For all other consumer types a matcher matching all attributes will be used.
The usage of a dedicated matcher can be enforced by the option <code>--matcher</code>.
`,
	}
}

func (o *Command) AddFlags(set *pflag.FlagSet) {
	set.StringVarP(&o.Type, "matcher", "m", "", "matcher type override")
	set.BoolVarP(&o.Sloppy, "sloppy", "s", false, "sloppy matching of consumer type")
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
	if o.Sloppy {
		fix := credentials.GuessConsumerType(o, o.Consumer.Type())
		if fix != o.Consumer.Type() {
			out.Outf(o, "Correcting consumer type to %q\n", fix)
			o.Consumer[credentials.ID_TYPE] = fix
		}
	}

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
