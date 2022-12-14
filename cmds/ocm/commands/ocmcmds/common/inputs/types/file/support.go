// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package file

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/mime"
)

type FileProcessSpec struct {
	cpi.MediaFileSpec
	Transformer func(ctx inputs.Context, inputDir string, data []byte) ([]byte, error)
}

func (s *FileProcessSpec) Validate(fldPath *field.Path, ctx inputs.Context, inputFilePath string) field.ErrorList {
	fileInfo, filePath, allErrs := s.MediaFileSpec.ValidateFile(fldPath, ctx, inputFilePath)
	if len(allErrs) == 0 {
		if !fileInfo.Mode().IsRegular() {
			pathField := fldPath.Child("path")
			allErrs = append(allErrs, field.Invalid(pathField, filePath, "no regular file"))
		}
	}
	return allErrs
}

func (s *FileProcessSpec) GetBlob(ctx inputs.Context, nv common.NameVersion, inputFilePath string) (accessio.TemporaryBlobAccess, string, error) {
	fs := ctx.FileSystem()
	inputInfo, inputPath, err := inputs.FileInfo(ctx, s.Path, inputFilePath)
	if err != nil {
		return nil, "", err
	}
	if inputInfo.IsDir() {
		return nil, "", fmt.Errorf("resource type is file but a directory was provided")
	}
	// otherwise just open the file
	var reader io.Reader
	inputBlob, err := fs.Open(inputPath)
	if err != nil {
		return nil, "", errors.Wrapf(err, "unable to read input blob from %q", inputPath)
	}
	reader = inputBlob

	var data []byte
	if s.Transformer != nil {
		data, err = io.ReadAll(inputBlob)
		inputBlob.Close()
		if err != nil {
			return nil, "", errors.Wrapf(err, "cannot read input file %s", inputPath)
		}
		dir, err := inputs.GetBaseDir(ctx.FileSystem(), inputFilePath)
		if err != nil {
			return nil, "", err
		}
		data, err = s.Transformer(ctx, dir, data)
		if err != nil {
			return nil, "", errors.Wrapf(err, "processing %a", inputPath)
		}
		reader = bytes.NewBuffer(data)
	}
	if !s.Compress() {
		if s.MediaType == "" {
			s.MediaType = mime.MIME_OCTET
		}
		if data == nil {
			inputBlob.Close()
			return accessio.TemporaryBlobAccessFor(accessio.BlobAccessForFile(s.MediaType, inputPath, fs)), "", nil
		}
		return accessio.TemporaryBlobAccessFor(accessio.BlobAccessForData(s.MediaType, data)), "", nil
	}

	temp, err := accessio.NewTempFile(fs, "", "compressed*.gzip")
	if err != nil {
		return nil, "", err
	}
	defer temp.Close()

	s.SetMediaTypeIfNotDefined(mime.MIME_GZIP)
	gw := gzip.NewWriter(temp.Writer())
	if _, err := io.Copy(gw, reader); err != nil {
		return nil, "", fmt.Errorf("unable to compress input file %q: %w", inputPath, err)
	}
	if err := gw.Close(); err != nil {
		return nil, "", fmt.Errorf("unable to close gzip writer: %w", err)
	}

	return temp.AsBlob(s.MediaType), "", nil
}

func Usage(head string) string {
	return `
` + head + `
The content is compressed if the <code>compress</code> field
is set to <code>true</code>.

This blob type specification supports the following fields: 
- **<code>path</code>** *string*

  This REQUIRED property describes the file path to the helm chart relative to the
  resource file location.
` + cpi.ProcessSpecUsage
}
