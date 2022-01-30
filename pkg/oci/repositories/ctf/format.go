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
	"github.com/gardener/ocm/pkg/common/accessobj"
	"github.com/gardener/ocm/pkg/oci/cpi"
	"github.com/gardener/ocm/pkg/oci/repositories/ctf/format"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

const ArtefactIndexFileName = format.ArtefactIndexFileName
const ArtefactsDirectoryName = format.ArtefactsDirectoryName

var accessObjectInfo = &accessobj.AccessObjectInfo{
	DescriptorFileName:       ArtefactIndexFileName,
	ObjectTypeName:           "repository",
	ElementDirectoryName:     ArtefactsDirectoryName,
	ElementTypeName:          "artefact",
	DescriptorHandlerFactory: NewStateHandler,
}

type Object = Repository

type FormatHandler struct {
	accessobj.FormatHandler
}

var (
	FormatDirectory = FormatHandler{accessobj.FormatDirectory}
	FormatTAR       = FormatHandler{accessobj.FormatTAR}
	FormatTGZ       = FormatHandler{accessobj.FormatTGZ}
)

func (h FormatHandler) Open(ctx cpi.Context, acc accessobj.AccessMode, path string, opts accessobj.Options) (*Object, error) {
	obj, err := h.FormatHandler.Open(accessObjectInfo, acc, path, opts)
	return _Wrap(ctx, obj, err)
}

func (h *FormatHandler) Create(ctx cpi.Context, path string, opts accessobj.Options, mode vfs.FileMode) (*Object, error) {
	obj, err := h.FormatHandler.Create(accessObjectInfo, path, opts, mode)
	return _Wrap(ctx, obj, err)
}

// WriteToFilesystem writes the current object to a filesystem
func (h *FormatHandler) Write(obj *Object, path string, opts accessobj.Options, mode vfs.FileMode) error {
	return h.FormatHandler.Write(obj.base, path, opts, mode)
}
