package describe

import (
	"github.com/mandelsoft/goutils/errors"
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/cachecmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/oci/attrs/cacheattr"
	"github.com/open-component-model/ocm/pkg/out"
)

var (
	Names = names.Cache
	Verb  = verbs.Describe
)

type Command struct {
	utils.BaseCommand
	cache accessio.BlobCache
}

// NewCommand creates a new artifact command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx)}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "",
		Short: "show OCI blob cache information",
		Long: `
Show details about the OCI blob cache (if given).
	`,
		Args: cobra.NoArgs,
		Example: `
$ ocm cache info
`,
	}
}

func (o *Command) Complete(args []string) error {
	o.cache = cacheattr.Get(o.Context)
	if o.cache == nil {
		return errors.Newf("no blob cache configured")
	}
	return nil
}

func (o *Command) Run() error {
	if r, ok := o.cache.(accessio.RootedCache); ok {
		path, fs := r.Root()
		out.Outf(o.Context, "Used cache directory %s [%s]\n", path, fs.Name())
	}

	if r, ok := o.cache.(accessio.CleanupCache); ok {
		cnt, _, _, size, _, _, err := r.Cleanup(nil, nil, true)
		if err != nil {
			return err
		}
		out.Outf(o.Context, "Total cache size %d entries [%.3f MB]\n", cnt, float64(size)/1024/1024)
	} else {
		out.Outf(o.Context, "Cache does not support more info\n")
	}

	return nil
}
