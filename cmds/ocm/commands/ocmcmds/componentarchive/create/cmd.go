package create

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/attrs/compatattr"
	"ocm.software/ocm/api/ocm/extensions/repositories/comparch"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/clisupport"
	"ocm.software/ocm/cmds/ocm/commands/common/options/formatoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/fileoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/schemaoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
var (
	//nolint:staticcheck // Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
	Names = names.ComponentArchive
	Verb  = verbs.Create
)

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
type Command struct {
	utils.BaseCommand

	providerattrs []string

	//nolint:staticcheck // Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
	Handler comparch.FormatHandler
	Force   bool
	Format  string

	Component      string
	Version        string
	Provider       string
	ProviderLabels metav1.Labels
	Labels         metav1.Labels
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx, formatoption.New(comparch.GetFormats()...), fileoption.NewCompArch(), schemaoption.New(compdesc.DefaultSchemeVersion))}, utils.Names(Names, names...)...)
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] <component> <version> --provider <provider-name> {--provider <label>=<value>} {<label>=<value>}",
		Args:  cobra.MinimumNArgs(2),
		Short: "(DEPRECATED) create new component archive",
		// this removes the command from the help output - https://github.com/open-component-model/ocm/issues/1242#issuecomment-2609312927
		// Deprecated: "Deprecated - use " + ocm.CommonTransportFormat + " instead",
		Example: `
$ ocm create componentarchive --file myfirst --provider acme.org --provider email=alice@acme.org acme.org/demo 1.0
`,
		Annotations: map[string]string{"ExampleCodeStyle": "bash"},
		Long: `
Create a new component archive. This might be either a directory prepared
to host component version content or a tar/tgz file (see option --type).

A provider must be specified, additional provider labels are optional.
`,
	}
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (o *Command) AddFlags(fs *pflag.FlagSet) {
	o.BaseCommand.AddFlags(fs)
	fs.BoolVarP(&o.Force, "force", "f", false, "remove existing content")
	fs.StringArrayVarP(&o.providerattrs, "provider", "p", nil, "provider attribute")
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (o *Command) Complete(args []string) error {
	var err error

	format := formatoption.From(o).Format
	o.Handler = comparch.GetFormat(format)
	if o.Handler == nil {
		return accessio.ErrInvalidFileFormat(format.String())
	}

	o.Component = args[0]
	o.Version = args[1]

	for _, a := range args[2:] {
		o.Labels, err = clisupport.AddParsedLabel(o.FileSystem(), o.Labels, a)
		if err != nil {
			return err
		}
	}
	for _, a := range o.providerattrs {
		if !strings.Contains(a, "=") {
			if o.Provider != "" {
				return fmt.Errorf("%s: provider name is already set (%s)", a, o.Provider)
			}
			o.Provider = a
			continue
		}
		o.ProviderLabels, err = clisupport.AddParsedLabel(o.FileSystem(), o.ProviderLabels, a)
		if err != nil {
			return err
		}
	}
	if o.Provider == "" {
		return fmt.Errorf("provider name missing")
	}
	return nil
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (o *Command) Run() error {
	mode := formatoption.From(o).Mode()
	fs := o.Context.FileSystem()
	fp := fileoption.From(o).Path
	if ok, err := vfs.Exists(fs, fp); ok || err != nil {
		if err != nil {
			return err
		}
		if o.Force {
			err = fs.RemoveAll(fp)
			if err != nil {
				return errors.Wrapf(err, "cannot remove old %q", fp)
			}
		}
	}
	obj, err := comparch.Create(o.Context.OCMContext(), accessobj.ACC_CREATE, fp, mode, o.Handler, fs)
	if err != nil {
		return err
	}
	desc := obj.GetDescriptor()
	desc.Metadata.ConfiguredVersion = schemaoption.From(o).Schema
	desc.Name = o.Component
	desc.Version = o.Version
	desc.Provider.Name = metav1.ProviderName(o.Provider)
	desc.Provider.Labels = o.ProviderLabels
	desc.Labels = o.Labels
	if !compatattr.Get(o.Context) {
		desc.CreationTime = metav1.NewTimestampP()
	}

	err = compdesc.Validate(desc)
	if err != nil {
		obj.Close()
		fs.RemoveAll(fp)
		return errors.Newf("invalid component info: %s", err)
	}
	err = obj.Close()
	if err != nil {
		fs.RemoveAll(fp)
	}
	return err
}
