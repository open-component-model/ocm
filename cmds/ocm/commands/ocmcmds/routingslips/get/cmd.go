package get

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm"
	common2 "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/cmds/ocm/commands/common/options/failonerroroption"
	ocmcommon "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/versionconstraintsoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/routingslips/common"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/output"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var (
	Names = names.RoutingSlips
	Verb  = verbs.Get
)

type Command struct {
	utils.BaseCommand

	Comp  string
	Slips []string
}

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{
		BaseCommand: utils.NewBaseCommand(ctx,
			NewOptions(),
			versionconstraintsoption.New(),
			repooption.New(),
			output.OutputOptions(outputs,
				failonerroroption.New(),
				lookupoption.New(),
			).OptimizeColumns(2).WithStatusCheck(check),
		),
	}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>]  <component> {<name>}",
		Args:  cobra.MinimumNArgs(1),
		Short: "get routings slips for a component version",
		Long: `
Get all or the selected routing slips for a component version specification.
`,
	}
}

func (o *Command) Complete(args []string) error {
	o.Comp = args[0]
	if len(args) > 1 {
		o.Slips = args[1:]
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

	opts := output.From(o)
	hdlr, err := common.NewTypeHandler(o.Context.OCM(), opts, repooption.From(o).Repository, session, []string{o.Comp}, common.WithVerification(From(o).Verify), common.OptionsFor(o))
	if err != nil {
		return err
	}
	specs := utils.StringElemSpecs(o.Slips...)
	return utils.HandleOutputs(opts, hdlr, specs...)
}

func check(opts *output.Options, e interface{}, err error) error {
	if failonerroroption.From(opts).Fail {
		elem := common.Elem(e)
		if elem.Error != "" {
			return fmt.Errorf("validation failed: for details see output")
		}
	}
	return err
}

////////////////////////////////////////////////////////////////////////////////

var outputs = output.NewOutputs(getRegular, output.Outputs{
	"wide": getWide,
}).AddManifestOutputs()

func getRegular(opts *output.Options) output.Output {
	return (&output.TableOutput{
		Headers: output.Fields("COMPONENT-VERSION", "NAME", "TYPE", "TIMESTAMP", "DESCRIPTION"),
		Options: opts,
		Mapping: mapGetRegularOutput,
	}).New()
}

func getWide(opts *output.Options) output.Output {
	return (&output.TableOutput{
		Headers: output.Fields("COMPONENT-VERSION", "NAME", "TYPE", "DIGEST", "PARENT", "TIMESTAMP", "LINKS", "DESCRIPTION"),
		Options: opts,
		Mapping: mapGetWideOutput,
	}).New()
}

func mapGetRegularOutput(e interface{}) interface{} {
	r := common.Elem(e)

	t := ""
	ts := ""
	desc := "Error: " + r.Error
	if r.HistoryEntry != nil {
		ts = r.HistoryEntry.Timestamp.String()
		t = r.HistoryEntry.Payload.GetType()
		if r.HistoryEntry.Payload != nil {
			desc = r.HistoryEntry.Payload.Describe(r.Component.ComponentVersion.GetContext())
		}
	}
	return []string{
		common2.VersionedElementKey(r.Component.ComponentVersion).String(),
		r.Slip,
		t,
		ts,
		desc,
	}
}

func mapGetWideOutput(e interface{}) interface{} {
	var links []string

	r := common.Elem(e)

	t := ""
	d := ""
	p := ""
	ts := ""
	desc := "Error: " + r.Error

	if r.HistoryEntry != nil {
		ts = r.HistoryEntry.Timestamp.String()
		t = r.HistoryEntry.Payload.GetType()
		d = r.HistoryEntry.Digest.Encoded()
		if len(d) > 8 {
			d = d[:8]
		}
		if r.HistoryEntry.Parent != nil {
			p = r.HistoryEntry.Parent.Encoded()
			if len(p) > 8 {
				p = p[:8]
			}
		}
		if r.HistoryEntry.Payload != nil {
			desc = r.HistoryEntry.Payload.Describe(r.Component.ComponentVersion.GetContext())
		}
		for _, l := range r.HistoryEntry.Links {
			if l.Name != "" && len(l.Digest.Encoded()) > 8 {
				links = append(links, fmt.Sprintf("%s@%s", l.Name, l.Digest.Encoded()[:8]))
			}
		}
	}
	return []string{
		common2.VersionedElementKey(r.Component.ComponentVersion).String(),
		r.Slip,
		t, d, p,
		ts,
		strings.Join(links, ","),
		desc,
	}
}
