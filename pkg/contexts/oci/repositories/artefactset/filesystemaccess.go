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

package artefactset

import (
	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi/support"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
)

type FileSystemBlobAccess struct {
	*accessobj.FileSystemBlobAccess
}

func NewFileSystemBlobAccess(access *accessobj.AccessObject) *FileSystemBlobAccess {
	return &FileSystemBlobAccess{accessobj.NewFileSystemBlobAccess(access)}
}

func (i *FileSystemBlobAccess) GetArtefact(access support.ArtefactSetContainerImpl, digest digest.Digest) (acc cpi.ArtefactAccess, err error) {

	v, err := access.View()
	if err != nil {
		return nil, err
	}
	defer v.Close()
	_, data, err := i.GetBlobData(digest)
	if err == nil {
		blob := accessio.BlobAccessForDataAccess("", -1, "", data)
		acc, err = support.NewArtefactForBlob(access, blob)
	}
	return
}

func (i *FileSystemBlobAccess) AddArtefactBlob(artefact cpi.Artefact) (cpi.BlobAccess, error) {
	blob, err := artefact.Blob()
	if err != nil {
		return nil, err
	}

	err = i.AddBlob(blob)
	if err != nil {
		return nil, err
	}
	return blob, nil
}
