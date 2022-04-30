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

package file

import (
	"compress/gzip"
	"fmt"
	"io"

	"github.com/open-component-model/ocm/cmds/ocm/clictx"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/mime"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type Spec struct {
	cpi.MediaFileSpec `json:",inline"`
}

var _ inputs.InputSpec = (*Spec)(nil)

func New(path, mediatype string, compress bool) *Spec {
	return &Spec{
		MediaFileSpec: cpi.NewMediaFileSpec(TYPE, path, mediatype, compress),
	}
}

func (s *Spec) Validate(fldPath *field.Path, ctx clictx.Context, inputFilePath string) field.ErrorList {
	fileInfo, filePath, allErrs := s.MediaFileSpec.ValidateFile(fldPath, ctx, inputFilePath)
	if len(allErrs) == 0 {
		if !fileInfo.Mode().IsRegular() {
			pathField := fldPath.Child("path")
			allErrs = append(allErrs, field.Invalid(pathField, filePath, "no regular file"))
		}
	}
	return allErrs
}

func (s *Spec) GetBlob(ctx clictx.Context, inputFilePath string) (accessio.TemporaryBlobAccess, string, error) {
	fs := ctx.FileSystem()
	inputInfo, inputPath, err := inputs.FileInfo(ctx, s.Path, inputFilePath)
	if err != nil {
		return nil, "", err
	}
	if inputInfo.IsDir() {
		return nil, "", fmt.Errorf("resource type is file but a directory was provided")
	}
	// otherwise just open the file
	inputBlob, err := fs.Open(inputPath)
	if err != nil {
		return nil, "", fmt.Errorf("unable to read input blob from %q: %w", inputPath, err)
	}

	if !s.Compress() {
		if s.MediaType == "" {
			s.MediaType = mime.MIME_OCTET
		}
		inputBlob.Close()
		return accessio.BlobNopCloser(accessio.BlobAccessForFile(s.MediaType, inputPath, fs)), "", nil
	}

	temp, err := accessio.NewTempFile(fs, "", "compressed*.gzip")
	if err != nil {
		return nil, "", err
	}
	defer temp.Close()

	s.SetMediaTypeIfNotDefined(mime.MIME_GZIP)
	gw := gzip.NewWriter(temp.Writer())
	if _, err := io.Copy(gw, inputBlob); err != nil {
		return nil, "", fmt.Errorf("unable to compress input file %q: %w", inputPath, err)
	}
	if err := gw.Close(); err != nil {
		return nil, "", fmt.Errorf("unable to close gzip writer: %w", err)
	}

	return temp.AsBlob(s.MediaType), "", nil
}

func (s *Spec) Usage() string {
	return `
- <code>file</code>

  The path must denote a file relative the the resources file.
  The content is compressed if the <code>compress</code> field
  is set to <code>true</code>.
`
}
