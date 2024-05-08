package clean

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/commands/cachecmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/oci/attrs/cacheattr"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/out"
	utils2 "github.com/open-component-model/ocm/pkg/utils"
)

var (
	Names = names.Cache
	Verb  = verbs.Clean
)

type Cache interface {
	accessio.BlobCache
	accessio.RootedCache
}

type Command struct {
	utils.BaseCommand
	cache accessio.CleanupCache

	duration string
	before   time.Time
	dryrun   bool
}

// NewCommand creates a new artifact command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx)}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "",
		Short: "cleanup oci blob cache",
		Long: `
Cleanup all blobs stored in oci blob cache (if given).
	`,
		Args: cobra.NoArgs,
		Example: `
$ ocm clean cache
`,
	}
}

func (o *Command) AddFlags(fs *pflag.FlagSet) {
	o.BaseCommand.AddFlags(fs)
	fs.StringVarP(&o.duration, "before", "b", "", "time since last usage")
	fs.BoolVarP(&o.dryrun, "dry-run", "s", false, "show size to be removed")
}

func (o *Command) Complete(args []string) error {
	c := cacheattr.Get(o.Context)
	if c == nil {
		return errors.Newf("no blob cache configured")
	}
	r, ok := c.(accessio.CleanupCache)
	if !ok {
		return errors.Newf("cache implementation does not support cleanup")
	}
	o.cache = r
	if o.duration != "" {
		if t, err := utils2.ParseDeltaTime(o.duration, true); err == nil {
			o.before = t
		} else {
			t, err := time.Parse(time.RFC3339, o.duration)
			if err != nil {
				t, err = time.Parse(o.duration, o.duration)
			}
			if err != nil {
				return fmt.Errorf("invalid lifetime %q", o.duration)
			}
			o.before = t
		}
	}
	return nil
}

func (o *Command) Run() error {

	cnt, ncnt, fcnt, size, nsize, fsize, err := o.cache.Cleanup(common.NewPrinter(o.Context.StdErr()), &o.before, o.dryrun)

	if err != nil {
		return err
	}
	if !o.before.IsZero() {
		if o.dryrun {
			out.Outf(o.Context, "Matching %d/%d entries [%.3f/%.3f MB]\n", cnt, ncnt+cnt, float64(size)/1024/1024, float64(size+nsize)/1024/1024)
		} else {
			out.Outf(o.Context, "Successfully deleted %d/%d entries [%.2f/%.3f MB]\n", cnt, ncnt+cnt, float64(size)/1024/1024, float64(size+nsize)/1024/1024)
		}
	} else {
		if o.dryrun {
			out.Outf(o.Context, "Would remove %d entries [%.3f MB]\n", cnt, float64(size)/1024/1024)
		} else {
			out.Outf(o.Context, "Successfully deleted %d entries [%.3f MB]\n", cnt, float64(size)/1024/1024)
		}
	}
	if fcnt > 0 {
		if o.dryrun {
			out.Outf(o.Context, "Failed to check %d entries [%.3f MB]\n", fcnt, float64(fsize)/1024/1024)
		} else {
			out.Outf(o.Context, "Failed to delete %d entries [%.3f MB]\n", fcnt, float64(fsize)/1024/1024)
		}
	}
	return nil
}
