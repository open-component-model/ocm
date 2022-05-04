// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package output

import (
	"strings"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/processing"
	. "github.com/open-component-model/ocm/pkg/out"
)

type StringOutput struct {
	ElementOutput
	linesep string
}

var _ Output = &StringOutput{}

func NewStringOutput(ctx Context, mapper processing.MappingFunction, linesep string) *StringOutput {
	return (&StringOutput{}).new(ctx, mapper, linesep)
}

func (this *StringOutput) new(ctx Context, mapper processing.MappingFunction, lineseperator string) *StringOutput {
	this.linesep = lineseperator
	this.ElementOutput.new(ctx, processing.Chain().Parallel(20).Map(mapper))
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
