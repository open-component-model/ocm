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

	"github.com/open-component-model/ocm/cmds/ocm/clictx"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/mime"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type Handler struct{}

var _ inputs.InputHandler = (*Handler)(nil)

func (h *Handler) Validate(fldPath *field.Path, ctx clictx.Context, input *inputs.BlobInput, inputFilePath string) field.ErrorList {
	allErrs := inputs.ForbidFilePattern(fldPath, input)
	path := fldPath.Child("path")
	if input.Path == "" {
		allErrs = append(allErrs, field.Required(path, "path is required for input"))
	} else {
		inputInfo, filePath, err := input.FileInfo(ctx, inputFilePath)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(path, filePath, err.Error()))
		}
		if !inputInfo.IsDir() {
			allErrs = append(allErrs, field.Invalid(path, filePath, "no directory"))
		}
	}
	return allErrs
}

func (h *Handler) GetBlob(ctx clictx.Context, input *inputs.BlobInput, inputFilePath string) (accessio.TemporaryBlobAccess, string, error) {
	fs := ctx.FileSystem()
	inputInfo, inputPath, err := input.FileInfo(ctx, inputFilePath)
	if !inputInfo.IsDir() {
		return nil, "", fmt.Errorf("resource type is dir but a file was provided")
	}

	opts := TarFileSystemOptions{
		IncludeFiles:   input.IncludeFiles,
		ExcludeFiles:   input.ExcludeFiles,
		PreserveDir:    input.PreserveDir != nil && *input.PreserveDir,
		FollowSymlinks: input.FollowSymlinks != nil && *input.FollowSymlinks,
	}

	temp, err := accessio.NewTempFile(fs, "", "resourceblob*.tgz")
	if err != nil {
		return nil, "", err
	}
	defer temp.Close()

	if input.Compress() {
		input.SetMediaTypeIfNotDefined(mime.MIME_GZIP)
		gw := gzip.NewWriter(temp.Writer())
		if err := TarFileSystem(fs, inputPath, gw, opts); err != nil {
			return nil, "", fmt.Errorf("unable to tar input artifact: %w", err)
		}
		if err := gw.Close(); err != nil {
			return nil, "", fmt.Errorf("unable to close gzip writer: %w", err)
		}
	} else {
		input.SetMediaTypeIfNotDefined(mime.MIME_TAR)
		if err := TarFileSystem(fs, inputPath, temp.Writer(), opts); err != nil {
			return nil, "", fmt.Errorf("unable to tar input artifact: %w", err)
		}
	}
	return temp.AsBlob(input.MediaType), "", nil
}

func (h *Handler) Usage() string {
	return `
- <code>dir</code>

  The path must denote a directory relative to the resources file, which is packed
  with tar and optionally compressed
  if the <code>compress</code> field is set to <code>true</code>. If the field
  <code>preserveDir</code> is set to true the directory itself is added to the tar.
  If the field <code>followSymLinks</code> is set to <code>true</code>, symbolic
  links are not packed but their targets files or folders.
  With the list fields <code>includeFiles</code> and <code>excludeFiles</code> it is 
  possible to specify which files should be included or excluded. The values are
  regular expression used to match relative file paths. If no inlcudes are specified
  all file not explicitly excluded are used.
`
}
