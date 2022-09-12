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

package artefacthdlr

import (
	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/data"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/processing"
	"github.com/open-component-model/ocm/pkg/common"
)

////////////////////////////////////////////////////////////////////////////////

var _ processing.TransformFunction = clean

// Clean is a processing chain cleaning and reordering closures.
var Clean = processing.Transform(clean)

func clean(iterable data.Iterable) data.Iterable {
	depth := map[digest.Digest]int{}
	tags := map[digest.Digest]string{}
	data := data.IndexedSliceAccess{}

	it := iterable.Iterator()

	for it.HasNext() {
		e := it.Next().(*Object)
		data.Add(e)
		l := len(e.History)
		blob, _ := e.Artefact.Blob()

		if l > depth[blob.Digest()] {
			depth[blob.Digest()] = l
			if e.Spec.Tag != nil {
				tags[blob.Digest()] = *e.Spec.Tag
			}
		}
	}

	output.Print(data, "clean in")
	for i := 0; i < len(data); i++ {
		e := data[i].(*Object)
		l := len(e.History)
		blob, _ := e.Artefact.Blob()
		dig := blob.Digest()
		d := depth[dig]
		if l == 0 && l < d && (e.Spec.Tag == nil || *e.Spec.Tag == tags[dig]) {
			j := i + 1
			prefix := append(e.History, common.NewNameVersion("", dig.String()))
			for ; j < len(data) && data[j].(*Object).History.HasPrefix(prefix); j++ {
			}
			data = append(data[:i], data[j:]...)
			i--
		}
	}
	output.Print(data, "clean reorg")
	return data
}
