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
	for _, a := range idx.Index {
		artefacts.AddArtefactInfo(&a)
	}
	return artefacts, nil
}

func (i StateHandler) Equivalent(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}
