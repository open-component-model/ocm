package output

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/mandelsoft/logging"
	"github.com/spf13/pflag"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/listformat"
	"ocm.software/ocm/cmds/ocm/common/options"
	"ocm.software/ocm/cmds/ocm/common/processing"
)

func From(o options.OptionSetProvider) *Options {
	var opts *Options
	if me, ok := o.(*Options); ok {
		return me
	}
	o.AsOptionSet().Get(&opts)
	return opts
}

func Selected(mode string) func(o options.OptionSetProvider) bool {
	return func(o options.OptionSetProvider) bool {
		return From(o).OutputMode == mode
	}
}

// StatusCheckFunction manipulates the processing status error according to
// the state of the given entry.
type StatusCheckFunction func(opts *Options, e interface{}, old error) error

type Options struct {
	options.OptionSet

	allColumns bool
	sort       []string

	Outputs          Outputs
	OutputMode       string
	Output           Output
	Sort             []string
	StatusCheck      StatusCheckFunction
	OptimizedColumns int
	FixedColums      int
	Context          clictx.Context // this context could be ocm context.
	Logging          logging.Context
	Session          ocm.Session
}

func OutputOptions(outputs Outputs, opts ...options.Options) *Options {
	return &Options{
		Outputs:   outputs,
		OptionSet: opts,
	}
}

func (o *Options) SortColumns(fields ...string) *Options {
	o.Sort = fields
	return o
}

func (o *Options) WithSession(s ocm.Session) *Options {
	o.Session = s
	return o
}

func (o *Options) OptimizeColumns(n int) *Options {
	o.OptimizedColumns = n
	return o
}

// WithStatusCheck provides the possibility to check every entry to
// influence the final out status.
func (o *Options) WithStatusCheck(f StatusCheckFunction) *Options {
	o.StatusCheck = f
	return o
}

func (o *Options) AdaptChain(errvar *error, chain processing.ProcessChain) processing.ProcessChain {
	if o.StatusCheck != nil {
		chain = processing.Map(func(e interface{}) interface{} {
			*errvar = o.StatusCheck(o, e, *errvar)
			return e
		}).Append(chain)
	}
	return chain
}

func (o *Options) LogContext() logging.Context {
	if o.Logging != nil {
		return o.Logging
	}
	return logging.DefaultContext()
}

func (o *Options) Options(proto options.Options) interface{} {
	return o.OptionSet.Options(proto)
}

func (o *Options) Get(proto interface{}) bool {
	return o.OptionSet.Get(proto)
}

func (o *Options) UseColumnOptimization() bool {
	return o.OptimizedColumns > 0 && !o.allColumns
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	s := ""
	if len(o.Outputs) > 1 {
		list := utils.StringMapKeys(o.Outputs)
		sep := ""
		for _, o := range list {
			if o != "" {
				s = fmt.Sprintf("%s%s%s", s, sep, o)
				sep = ", "
			}
		}
		fs.StringVarP(&o.OutputMode, "output", "o", "", fmt.Sprintf("output mode (%s)", s))
	}

	// TODO: not the best solution to instantiate all possible outputs to figure out, whether sort fields
	// are available or not
	for _, out := range o.Outputs {
		if _, ok := out(o).(SortFields); ok {
			fs.StringArrayVarP(&o.sort, "sort", "s", nil, "sort fields")
			break
		}
	}

	if o.OptimizedColumns > 0 {
		fs.BoolVarP(&o.allColumns, "all-columns", "", false, "show all table columns")
	}
	o.OptionSet.AddFlags(fs)
}

func (o *Options) Configure(ctx clictx.Context) error {
	o.Context = ctx

	// process sub options first, to assure that output options are available for output
	// mode creation
	err := o.OptionSet.ProcessOnOptions(options.CompleteOptionsWithCLIContext(ctx))
	if err != nil {
		return err
	}

	if f := o.Outputs[o.OutputMode]; f == nil {
		return errors.ErrInvalid("output mode", o.OutputMode)
	} else {
		o.Output = f(o)
	}

	var avail sliceutils.OrderedSlice[string]

	if s, ok := o.Output.(SortFields); ok {
		avail = s.GetSortFields()
	}
	if len(o.sort) > 0 {
		var fields []string
		for _, f := range o.sort {
			split := strings.Split(f, ",")
			for _, s := range split {
				s = strings.TrimSpace(s)
				if s != "" {
					if avail.Contains(s) {
						fields = append(fields, s)
					} else {
						return errors.ErrInvalid("sort field", s)
					}
				}
			}
		}
		o.Sort = fields
	}
	return nil
}

func (o *Options) CompleteAll(ctx clictx.Context) error {
	err := o.Configure(ctx)
	if err == nil {
		err = o.OptionSet.ProcessOnOptions(options.CompleteOptionsWithCLIContext(ctx))
	}
	if err != nil {
		return err
	}
	return err
}

func (o *Options) Usage() string {
	s := o.OptionSet.Usage()

	if len(o.Outputs) > 1 {
		s += `
With the option <code>--output</code> the output mode can be selected.
The following modes are supported:
` + listformat.FormatList(o.OutputMode, utils.StringMapKeys(o.Outputs)...)
	}
	return s
}

////////////////////////////////////////////////////////////////////////////////

func OutputModeCondition(opts *Options, mode string) options.Condition {
	return options.Flag(opts.OutputMode == mode)
}
