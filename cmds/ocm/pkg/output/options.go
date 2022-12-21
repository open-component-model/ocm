// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package output

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mandelsoft/logging"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/utils"
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

type Options struct {
	options.OptionSet

	Outputs     Outputs
	OutputMode  string
	Output      Output
	Sort        []string
	FixedColums int
	Context     clictx.Context // this context could be ocm context.
	Logging     logging.Context
}

func OutputOptions(outputs Outputs, opts ...options.Options) *Options {
	return &Options{
		Outputs:   outputs,
		OptionSet: opts,
	}
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

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	s := ""
	if len(o.Outputs) > 1 {
		list := []string{}
		for o := range o.Outputs {
			list = append(list, o)
		}
		sort.Strings(list)
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
			fs.StringArrayVarP(&o.Sort, "sort", "s", nil, "sort fields")
			break
		}
	}

	o.OptionSet.AddFlags(fs)
}

func (o *Options) Complete(ctx clictx.Context) error {
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

	var avail utils.StringSlice

	var fields []string

	if s, ok := o.Output.(SortFields); ok {
		avail = s.GetSortFields()
	}
	for _, f := range o.Sort {
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
	return nil
}

func (o *Options) CompleteAll(ctx clictx.Context) error {
	err := o.Complete(ctx)
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
`
		list := []string{}
		for o := range o.Outputs {
			list = append(list, o)
		}
		sort.Strings(list)
		for _, m := range list {
			if m != "" {
				s += " - " + m + "\n"
			}
		}
	}
	return s
}

////////////////////////////////////////////////////////////////////////////////

func OutputModeCondition(opts *Options, mode string) options.Condition {
	return options.Flag(opts.OutputMode == mode)
}
