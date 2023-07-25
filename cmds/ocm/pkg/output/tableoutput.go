// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package output

import (
	"strings"

	. "github.com/open-component-model/ocm/v2/cmds/ocm/pkg/processing"

	"github.com/open-component-model/ocm/v2/cmds/ocm/pkg/data"
	"github.com/open-component-model/ocm/v2/pkg/errors"
	"github.com/open-component-model/ocm/v2/pkg/out"
)

type SortFields interface {
	GetSortFields() []string
}

type TableOutput struct {
	Headers []string
	Options *Options
	Chain   ProcessChain
	Mapping MappingFunction
}

var _ SortFields = (*TableOutput)(nil)

func (t *TableOutput) New() *TableProcessingOutput {
	chain := t.Chain
	if chain == nil {
		chain = Map(t.Mapping)
	} else {
		chain = chain.Map(t.Mapping)
	}
	return NewProcessingTableOutput(t.Options, chain, t.Headers...)
}

func (this *TableOutput) GetSortFields() []string {
	return this.Headers[this.Options.FixedColums:]
}

type TableProcessingOutput struct {
	ElementOutput
	header []string
	opts   *Options
}

var (
	_ Output     = (*TableProcessingOutput)(nil)
	_ SortFields = (*TableProcessingOutput)(nil)
)

func NewProcessingTableOutput(opts *Options, chain ProcessChain, header ...string) *TableProcessingOutput {
	return (&TableProcessingOutput{}).new(opts, chain, header)
}

func (this *TableProcessingOutput) new(opts *Options, chain ProcessChain, header []string) *TableProcessingOutput {
	this.header = header
	this.ElementOutput.new(opts.LogContext(), opts.Context, chain)
	this.opts = opts
	return this
}

func (this *TableProcessingOutput) GetSortFields() []string {
	return this.header[this.opts.FixedColums:]
}

func (this *TableProcessingOutput) Out() error {
	lines := [][]string{this.header}

	sort := this.opts.Sort
	slice := data.IndexedSliceAccess(data.Slice(this.Elems))
	if len(slice) == 0 {
		out.Out(this.Context, "no elements found\n")
		return nil
	}
	if sort != nil {
		cols := make([]string, len(this.header))
		idxs := map[string]int{}
		for i, n := range this.header {
			cols[i] = strings.TrimPrefix(strings.ToLower(n), "-")
			idxs[cols[i]] = i
		}
		for _, k := range sort {
			key, n := SelectBest(strings.ToLower(k), cols...)
			if key == "" {
				return errors.Newf("unknown field '%s'", k)
			}
			if n < this.opts.FixedColums {
				return errors.Newf("field '%s' not possible", k)
			}
			cmp := compareColumn(idxs[key])
			if this.opts.FixedColums > 0 {
				sortFixed(this.opts.FixedColums, slice, cmp)
			} else {
				slice.Sort(cmp)
			}
		}
	}

	FormatTable(this.Context, "", append(lines, data.StringArraySlice(slice)...))
	return nil
}

func compareColumn(c int) CompareFunction {
	return func(a interface{}, b interface{}) int {
		aa := a.([]string)
		ab := b.([]string)
		if len(aa) > c && len(ab) > c {
			return strings.Compare(aa[c], ab[c])
		}
		return len(aa) - len(ab)
	}
}

func sortFixed(fixed int, slice data.IndexedSliceAccess, cmp CompareFunction) {
	keys := [][]string{}
	views := [][]int{}
lineloop:
	for l, e := range slice {
		line := e.([]string)
	keyloop:
		for k, v := range keys {
			for i := 0; i < fixed; i++ {
				if v[i] != line[i] {
					continue keyloop
				}
			}
			views[k] = append(views[k], l)
			continue lineloop
		}
		keys = append(keys, line[:fixed])
		views = append(views, []int{l})
	}
	for _, v := range views {
		data.SortView(slice, v, cmp)
	}
}
