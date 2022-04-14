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
	"fmt"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/output/out"
)

type AttributeSet struct {
	attrs [][]string
}

func NewAttributeSet() *AttributeSet {
	a := &AttributeSet{}
	a.ResetAttributes()
	return a
}

func (this *AttributeSet) ResetAttributes() {
	this.attrs = [][]string{[]string{}}
}

func (this *AttributeSet) Attribute(name, value string) {
	this.attrs = append(this.attrs, []string{name + ":", value})
}

func (this *AttributeSet) Attributef(name, f string, args ...interface{}) {
	this.attrs = append(this.attrs, []string{name + ":", fmt.Sprintf(f, args...)})
}

func (this *AttributeSet) PrintAttributes(ctx out.Context) {
	FormatTable(ctx, "", this.attrs)
}
