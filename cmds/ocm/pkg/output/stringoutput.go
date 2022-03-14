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
	"context"
	"fmt"
	"strings"

	"github.com/gardener/ocm/cmds/ocm/pkg/data"
)

type StringOutput struct {
	ElementOutput
	linesep string
}

var _ Output = &StringOutput{}

func NewStringOutput(mapper data.MappingFunction, linesep string) *StringOutput {
	return (&StringOutput{}).new(mapper, linesep)
}

func (this *StringOutput) new(mapper data.MappingFunction, lineseperator string) *StringOutput {
	this.linesep = lineseperator
	this.ElementOutput.new(data.Chain().Parallel(20).Map(mapper))
	return this
}

func (this *StringOutput) Out(ctx *context.Context) error {
	var err error = nil
	i := this.Elems.Iterator()
	for i.HasNext() {
		switch cfg := i.Next().(type) {
		case error:
			err = cfg
			if this.linesep == "" {
				fmt.Printf("Error: %s\n", err)
			} else {
				fmt.Printf("%s\nError: %s\n", this.linesep, err)
			}
		case string:
			if cfg != "" {
				if this.linesep != "" {
					if !strings.HasPrefix(cfg, this.linesep+"\n") {
						fmt.Println(this.linesep)
					}
				}
				fmt.Println(cfg)
			}
		}
	}
	return err
}
