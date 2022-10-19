// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package output

import (
	"strings"

	. "github.com/open-component-model/ocm/pkg/out"

	"github.com/mandelsoft/logging"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/processing"
)

type StringOutput struct {
	ElementOutput
	linesep string
}

var _ Output = &StringOutput{}

func NewStringOutput(log logging.Context, ctx Context, mapper processing.MappingFunction, linesep string) *StringOutput {
	return (&StringOutput{}).new(log, ctx, mapper, linesep)
}

func (this *StringOutput) new(log logging.Context, ctx Context, mapper processing.MappingFunction, lineseperator string) *StringOutput {
	this.linesep = lineseperator
	this.ElementOutput.new(log, ctx, processing.Chain(log).Parallel(20).Map(mapper))
	return this
}

func (this *StringOutput) Out() error {
	var err error = nil
	i := this.Elems.Iterator()
	for i.HasNext() {
		switch cfg := i.Next().(type) {
		case error:
			err = cfg
			if this.linesep == "" {
				Error(this.Context, err.Error())
			} else {
				Errf(this.Context, "%s\nError: %s\n", this.linesep, err)
			}
		case string:
			if cfg != "" {
				if this.linesep != "" {
					if !strings.HasPrefix(cfg, this.linesep+"\n") {
						Outln(this.Context, this.linesep)
					}
				}
				Outln(this.Context, cfg)
			}
		}
	}
	return err
}
