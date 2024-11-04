package execute

import (
	"encoding/json"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/datacontext/action"
	"ocm.software/ocm/api/datacontext/action/api"
	utils2 "ocm.software/ocm/api/utils"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/out"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var (
	Names = names.Action
	Verb  = verbs.Execute
)

type Command struct {
	utils.BaseCommand

	Name        string
	Spec        action.ActionSpec
	OutputMode  string
	MatcherType string

	Matcher  credentials.IdentityMatcher
	Consumer credentials.ConsumerIdentity
}

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(
		&Command{
			BaseCommand: utils.NewBaseCommand(ctx),
		},
		utils.Names(Names, names...)...,
	)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] <action spec> {<cred>=<value>}",
		Short: "execute an action",
		Args:  cobra.MinimumNArgs(1),
		Long: `
Execute an action extension for a given action specification. The specification
should be a JSON or YAML argument.

Additional properties settings can be used to describe a consumer id
to retrieve credentials for.

The following actions are supported:
` + api.Usage(api.DefaultRegistry()),
		Example: `
$ ocm execute action '{ "type": "oci.repository.prepare/v1", "hostname": "...", "repository": "..."}'
`,
		Annotations: map[string]string{"ExampleCodeStyle": "bash"},
	}
}

func (o *Command) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.MatcherType, "matcher", "m", "", "matcher type override")
	fs.StringVarP(&o.OutputMode, "output", "o", "json", "output mode (json, yaml)")
	fs.StringVarP(&o.Name, "name", "n", "", "action name (overrides type in specification)")
}

func (o *Command) Complete(args []string) error {
	var err error

	data := []byte(args[0])
	if strings.HasPrefix(args[0], "@") {
		data, err = utils2.ResolveData(args[0][1:], o.FileSystem())
		if err != nil {
			return err
		}
	}

	if o.OutputMode != "json" && o.OutputMode != "yaml" {
		return errors.Wrapf(err, "invalid output mode %q", o.OutputMode)
	}

	var un runtime.UnstructuredVersionedTypedObject

	err = runtime.DefaultYAMLEncoding.Unmarshal(data, &un)
	if err != nil {
		return errors.Wrapf(err, "invalid action spec")
	}
	if o.Name != "" {
		un.SetType(o.Name)
	} else {
		o.Name = un.GetKind()
	}

	data, err = json.Marshal(&un)
	if err != nil {
		return errors.Wrapf(err, "cannot marshal final spec")
	}
	o.Spec, err = o.GetActions().GetActionTypes().DecodeActionSpec(data, runtime.DefaultYAMLEncoding)

	if o.MatcherType != "" {
		m := o.CredentialsContext().ConsumerIdentityMatchers().Get(o.MatcherType)
		if m == nil {
			return errors.ErrUnknown("identity matcher", o.MatcherType)
		}
		o.Matcher = m
	}
	o.Consumer = credentials.ConsumerIdentity{}
	for _, s := range args[1:] {
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

	return err
}

func (o *Command) Run() error {
	var creds common.Properties

	if len(o.Consumer) > 0 {
		c, err := credentials.RequiredCredentialsForConsumer(o.CredentialsContext(), o.Consumer, o.Matcher)
		if err != nil {
			return err
		}
		creds = c.Properties()
		out.Outf(o, "Using credentials\n")
	}

	out.Outf(o, "Executing action %s...\n", o.Name)
	r, err := o.Context.OCMContext().GetActions().Execute(o.Spec, creds)
	if err != nil {
		return errors.Wrapf(err, "execution failed")
	}

	var data []byte

	if o.OutputMode == "json" {
		data, err = runtime.DefaultJSONEncoding.Marshal(r)
	} else {
		data, err = runtime.DefaultYAMLEncoding.Marshal(r)
	}
	if err != nil {
		return errors.Wrapf(err, "cannot marshal result")
	}
	out.Outf(o, "%s", string(data))
	return nil
}
