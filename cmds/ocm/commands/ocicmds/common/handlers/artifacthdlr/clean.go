// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package artifacthdlr

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
		blob, _ := e.Artifact.Blob()

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
		blob, _ := e.Artifact.Blob()
		dig := blob.Digest()
		d := depth[dig]
		if l == 0 && l < d && (e.Spec.Tag == nil || *e.Spec.Tag == tags[dig]) {
			j := i + 1
			prefix := e.History.Append(common.NewNameVersion("", dig.String()))
			for ; j < len(data) && data[j].(*Object).History.HasPrefix(prefix); j++ {
			}
			data = append(data[:i], data[j:]...)
			i--
		}
	}
	output.Print(data, "clean reorg")
	return data
}
