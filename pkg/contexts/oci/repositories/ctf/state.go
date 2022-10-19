// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ctf

import (
	"fmt"
	"reflect"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf/index"
)

type StateHandler struct {
	fs vfs.FileSystem
}

var _ accessobj.StateHandler = &StateHandler{}

func NewStateHandler(fs vfs.FileSystem) accessobj.StateHandler {
	return &StateHandler{fs}
}

func (i StateHandler) Initial() interface{} {
	return index.NewRepositoryIndex()
}

func (i StateHandler) Encode(d interface{}) ([]byte, error) {
	return index.Encode(d.(*index.RepositoryIndex).GetDescriptor())
}

func (i StateHandler) Decode(data []byte) (interface{}, error) {
	idx, err := index.Decode(data)
	if err != nil {
		return nil, fmt.Errorf("unable to parse artefact index read from %s: %w", ArtefactIndexFileName, err)
	}
	if idx.SchemaVersion != index.SchemaVersion {
		return nil, fmt.Errorf("unknown schema version %d for artefact index %s", index.SchemaVersion, ArtefactIndexFileName)
	}

	artefacts := index.NewRepositoryIndex()
	for i := range idx.Index {
		artefacts.AddArtefactInfo(&idx.Index[i])
	}
	return artefacts, nil
}

func (i StateHandler) Equivalent(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}
