package fileoption

import (
	"archive/tar"
	"fmt"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/pflag"

	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/cobrautils/flag"
	"ocm.software/ocm/api/utils/compression"
	"ocm.software/ocm/cmds/ocm/common/options"
)

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func NewCompArch() *Option {
	return New("component-archive")
}

func New(def string, us ...interface{}) *Option {
	usage := fmt.Sprint(us...)
	if usage == "" {
		usage = "target file/directory"
	}
	return &Option{def: def, usage: usage}
}

type Option struct {
	flag  *pflag.Flag
	def   string
	usage string
	Path  string
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	o.flag = flag.StringVarPF(fs, &o.Path, "file", "F", o.def, o.usage)
}

func (o *Option) IsSet() bool {
	return o.flag.Changed
}

// GetPath return a path depending on the option setting and the first argument.
// if the option is not set and the first argument denotes a path to a directory or tar file,
// the first argument if chosen as path.
func (o *Option) GetPath(args []string, fss ...vfs.FileSystem) (string, []string) {
	if o.IsSet() || len(args) == 0 {
		return o.Path, args
	}

	fs := utils.FileSystem(fss...)
	if ok, err := vfs.Exists(fs, args[0]); !ok || err != nil {
		return o.Path, args
	}
	// ignored: I can't rewrite this whole segment.
	if ok, _ := vfs.IsDir(fs, args[0]); ok {
		return args[0], args[1:]
	}

	file, err := fs.Open(args[0])
	if err != nil {
		return o.Path, args
	}
	defer file.Close()
	r, _, err := compression.AutoDecompress(file)
	if err != nil {
		return o.Path, args
	}
	_, err = tar.NewReader(r).Next()
	if err != nil {
		return o.Path, args
	}
	return args[0], args[1:]
}
