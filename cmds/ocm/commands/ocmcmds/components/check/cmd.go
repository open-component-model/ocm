package check

import (
	"fmt"

	"github.com/mandelsoft/goutils/optionutils"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/json"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/ocmutils/check"
	utils2 "ocm.software/ocm/api/utils"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/cmds/ocm/commands/common/options/failonerroroption"
	ocmcommon "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/handlers/comphdlr"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/output"
	"ocm.software/ocm/cmds/ocm/common/processing"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var (
	Names = names.Components
	Verb  = verbs.Check
)

type Command struct {
	utils.BaseCommand

	Refs []string
}

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(
		&Command{
			BaseCommand: utils.NewBaseCommand(ctx,
				repooption.New(),
				output.OutputOptions(outputs,
					failonerroroption.New(),
					NewOption(),
				),
			),
		},
		utils.Names(Names, names...)...,
	)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] {<component-reference>}",
		Short: "Check completeness of a component version in an OCM repository",
		Long: `
This command checks, whether component versions are completely contained
in an OCM repository with all its dependent component references.
`,
		Example: `
$ ocm check componentversion ghcr.io/open-component-model/ocm//ocm.software/ocmcli:0.17.0
$ ocm check componentversion --repo OCIRegistry::ghcr.io/open-component-model/ocm ocm.software/ocmcli:0.17.0
`,
		Annotations: map[string]string{"ExampleCodeStyle": "bash"},
	}
}

func (o *Command) Complete(args []string) error {
	o.Refs = args
	if len(args) == 0 && repooption.From(o).Spec == "" {
		return fmt.Errorf("a repository or at least one argument that defines the reference is needed")
	}
	return nil
}

func (o *Command) Run() error {
	session := ocm.NewSession(nil)
	defer session.Close()

	err := o.ProcessOnOptions(ocmcommon.CompleteOptionsWithSession(o, session))
	if err != nil {
		return err
	}
	handler := comphdlr.NewTypeHandler(o.Context.OCM(), session, repooption.From(o).Repository, comphdlr.OptionsFor(o))
	err = utils.HandleArgs(output.From(o), handler, o.Refs...)
	if err != nil {
		return err
	}
	return failonerroroption.From(o).ActivatedError()
}

////////////////////////////////////////////////////////////////////////////////

var outputs = output.NewOutputs(OutputFactory(mapRegularOutput), output.Outputs{
	"wide": OutputFactory(mapWideOutput, "MISSING", "NON-LOCAL"),
}).AddChainedManifestOutputs(NewAction)

func OutputFactory(fmt processing.MappingFunction, wide ...string) output.OutputFactory {
	return func(opts *output.Options) output.Output {
		return (&output.TableOutput{
			Headers: output.Fields("COMPONENT", "VERSION", "STATUS", "ERROR", wide),
			Options: opts,
			Chain:   NewAction(opts),
			Mapping: fmt,
		}).New()
	}
}

func mapRegularOutput(e interface{}) interface{} {
	p := e.(*Entry)

	err := ""
	if p.Error != nil {
		err = p.Error.Error()
	}
	return []string{p.ComponentVersion.GetName(), p.ComponentVersion.GetVersion(), p.Status, err}
}

func mapWideOutput(e interface{}) interface{} {
	p := e.(*Entry)

	line := mapRegularOutput(e).([]string)
	if p.Results.IsEmpty() {
		return append(line, "")
	}

	mmsg := ""
	amsg := ""
	if len(p.Results.Missing) > 0 {
		missing := map[string]string{}
		for id, m := range p.Results.Missing {
			sep := "["
			d := ""
			for _, id := range m[:len(m)-1] {
				d = d + sep + id.String()
				sep = "->"
			}
			missing[id.String()] = d + "]"
		}
		sep := ""
		for _, k := range utils2.StringMapKeys(missing) {
			mmsg += sep + k + missing[k]
			sep = ", "
		}
	}

	if len(p.Results.Resources) > 0 {
		sep := "RSC("
		for _, r := range p.Results.Resources {
			amsg = fmt.Sprintf("%s%s%s", amsg, sep, r.String())
			sep = ","
		}
		amsg += ")"
	}
	if len(p.Results.Sources) > 0 {
		sep := "SRC("
		for _, r := range p.Results.Sources {
			amsg = fmt.Sprintf("%s%s%s", amsg, sep, r.String())
			sep = ","
		}
		amsg += ")"
	}

	return append(line, mmsg, amsg)
}

////////////////////////////////////////////////////////////////////////////////

type CheckResult = check.Result

type Entry struct {
	Status           string             `json:"status"`
	ComponentVersion common.NameVersion `json:"componentVersion"`
	Results          *CheckResult       `json:",inline"` // does not work
	Error            error              `json:"error,omitempty"`
}

func (n Entry) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{}
	if n.Results != nil {
		data, err := json.Marshal(n.Results)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(data, &m)
		if err != nil {
			return nil, err
		}
	}
	m["status"] = n.Status
	m["componentVersion"] = n.ComponentVersion
	if n.Error != nil {
		m["error"] = n.Error
	}
	return json.Marshal(m)
}

type action struct {
	erropt  *failonerroroption.Option
	options *check.Options
}

func NewAction(opts *output.Options) processing.ProcessChain {
	return comphdlr.Sort.Map((&action{
		erropt:  failonerroroption.From(opts),
		options: optionutils.EvalOptions(check.Option(From(opts))),
	}).Map)
}

func (a *action) Map(in interface{}) interface{} {
	i := in.(*comphdlr.Object)
	o := &Entry{
		ComponentVersion: common.VersionedElementKey(i.ComponentVersion),
	}
	status := ""
	o.Results, o.Error = a.options.For(i.ComponentVersion)
	if o.Error != nil {
		status = ",Error"
		a.erropt.AddError(o.Error)
	}
	if !o.Results.IsEmpty() {
		if len(o.Results.Missing) > 0 {
			a.erropt.AddError(fmt.Errorf("incomplete component version %s", common.VersionedElementKey(i.ComponentVersion)))
			status += ",Incomplete"
		}
		if len(o.Results.Sources) > 0 || len(o.Results.Resources) > 0 {
			if len(o.Results.Resources) > 0 {
				status += ",Resources"
				a.erropt.AddError(fmt.Errorf("version %s with non-local resources", common.VersionedElementKey(i.ComponentVersion)))
			}
			if len(o.Results.Sources) > 0 {
				status += ",Sources"
				a.erropt.AddError(fmt.Errorf("version %s with non-local sources", common.VersionedElementKey(i.ComponentVersion)))
			}
		}
	}
	if status != "" {
		o.Status = status[1:]
	} else {
		o.Status = "OK"
	}
	return o
}
