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

package directory

import (
	"compress/gzip"
	"fmt"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/utils/tarutils"
)

type Spec struct {
	cpi.MediaFileSpec `json:",inline"`
	// PreserveDir defines that the directory specified in the Path field should be included in the blob.
	// Only supported for Type dir.
	PreserveDir *bool `json:"preserveDir,omitempty"`
	// IncludeFiles is a list of shell file name patterns that describe the files that should be included.
	// If nothing is defined all files are included.
	// Only relevant for blobinput type "dir".
	IncludeFiles []string `json:"includeFiles,omitempty"`
	// ExcludeFiles is a list of shell file name patterns that describe the files that should be excluded from the resulting tar.
	// Excluded files always overwrite included files.
	// Only relevant for blobinput type "dir".
	ExcludeFiles []string `json:"excludeFiles,omitempty"`
	// FollowSymlinks configures to follow and resolve symlinks when a directory is tarred.
	// This options will include the content of the symlink directly in the tar.
	// This option should be used with care.
	FollowSymlinks *bool `json:"followSymlinks,omitempty"`
}

var _ inputs.InputSpec = (*Spec)(nil)

func New(path, mediatype string, compress bool) *Spec {
	return &Spec{
		MediaFileSpec:  cpi.NewMediaFileSpec(TYPE, path, mediatype, compress),
		PreserveDir:    nil,
		IncludeFiles:   nil,
		ExcludeFiles:   nil,
		FollowSymlinks: nil,
	}
}

func (s *Spec) Validate(fldPath *field.Path, ctx inputs.Context, inputFilePath string) field.ErrorList {
	fileInfo, filePath, allErrs := s.MediaFileSpec.ValidateFile(fldPath, ctx, inputFilePath)
	if len(allErrs) == 0 {
		if !fileInfo.Mode().IsDir() {
			pathField := fldPath.Child("path")
			allErrs = append(allErrs, field.Invalid(pathField, filePath, "no directory"))
		}
	}
	return allErrs
}

func (s *Spec) GetBlob(ctx inputs.Context, nv common.NameVersion, inputFilePath string) (accessio.TemporaryBlobAccess, string, error) {
	fs := ctx.FileSystem()
	inputInfo, inputPath, err := inputs.FileInfo(ctx, s.Path, inputFilePath)
	if err != nil {
		return nil, "", fmt.Errorf("resource dir %s: %w", inputFilePath, err)
	}
	if !inputInfo.IsDir() {
		return nil, "", fmt.Errorf("resource type is dir but a file was provided")
	}

	opts := tarutils.TarFileSystemOptions{
		IncludeFiles:   s.IncludeFiles,
		ExcludeFiles:   s.ExcludeFiles,
		PreserveDir:    s.PreserveDir != nil && *s.PreserveDir,
		FollowSymlinks: s.FollowSymlinks != nil && *s.FollowSymlinks,
	}

	temp, err := accessio.NewTempFile(fs, "", "resourceblob*.tgz")
	if err != nil {
		return nil, "", err
	}
	defer temp.Close()

	if s.Compress() {
		s.SetMediaTypeIfNotDefined(mime.MIME_TGZ)
		gw := gzip.NewWriter(temp.Writer())
		if err := tarutils.TarFileSystem(fs, inputPath, gw, opts); err != nil {
			return nil, "", fmt.Errorf("unable to tar input artifact: %w", err)
		}
		if err := gw.Close(); err != nil {
			return nil, "", fmt.Errorf("unable to close gzip writer: %w", err)
		}
	} else {
		s.SetMediaTypeIfNotDefined(mime.MIME_TAR)
		if err := tarutils.TarFileSystem(fs, inputPath, temp.Writer(), opts); err != nil {
			return nil, "", fmt.Errorf("unable to tar input artifact: %w", err)
		}
	}
	return temp.AsBlob(s.MediaType), "", nil
}
