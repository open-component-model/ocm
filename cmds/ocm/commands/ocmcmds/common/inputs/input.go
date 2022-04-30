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

package inputs

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/open-component-model/ocm/cmds/ocm/clictx"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/mime"
)

// BlobInput defines a local resource input that should be added to the component descriptor and
// to the resource's access.
type BlobInput struct {
	// Type defines the input type of the blob to be added.
	// Note that a input blob of type "dir" is automatically tarred.
	Type string `json:"type"`
	// MediaType is the mediatype of the defined file that is also added to the oci layer.
	// Should be a custom media type in the form of "application/vnd.<mydomain>.<my description>"
	MediaType string `json:"mediaType,omitempty"`
	// Path is the path that points to the blob to be added.
	Path string `json:"path"`
	// CompressWithGzip defines that the blob should be automatically compressed using gzip.
	CompressWithGzip *bool `json:"compress,omitempty"`
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

// Compress returns if the blob should be compressed using gzip.
func (input BlobInput) Compress() bool {
	if input.CompressWithGzip == nil {
		return mime.IsGZip(input.MediaType)
	}
	return *input.CompressWithGzip
}

// SetMediaTypeIfNotDefined sets the media type of the input blob if its not defined
func (input *BlobInput) SetMediaTypeIfNotDefined(mediaType string) {
	if len(input.MediaType) != 0 {
		return
	}
	input.MediaType = mediaType
}

// GetBlob provides a BlobAccess for the actual input.
func (input *BlobInput) GetBlob(ctx clictx.Context, inputFilePath string) (accessio.TemporaryBlobAccess, string, error) {
	hdlr := Default.Get(input.Type)
	if hdlr == nil {
		return nil, "", fmt.Errorf("unknown input type %q", input.Type)
	}
	return hdlr.GetBlob(ctx, input, inputFilePath)
}

func (input *BlobInput) Validate(fldPath *field.Path, ctx clictx.Context, inputFilePath string) field.ErrorList {
	if input == nil {
		return nil
	}
	allErrs := field.ErrorList{}
	path := fldPath.Child("type")
	if input.Type == "" {
		allErrs = append(allErrs, field.Required(path, "input type required"))
	} else {
		hdlr := Default.Get(input.Type)
		if hdlr == nil {
			allErrs = append(allErrs, field.NotSupported(path, input.Type, Default.KnownTypes()))
		} else {
			allErrs = append(allErrs, hdlr.Validate(fldPath, ctx, input, inputFilePath)...)
		}
	}
	return allErrs
}

func (input *BlobInput) Evaluate(scheme InputTypeScheme) (InputSpec, error) {
	return nil, nil
}
