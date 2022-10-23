// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cpi

import (
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/options"
	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type PathSpec struct {
	runtime.ObjectVersionedType `json:",inline"`

	// Path is a file or repository path
	Path string `json:"path"`
}

func NewPathSpec(typ, path string) PathSpec {
	return PathSpec{
		ObjectVersionedType: runtime.ObjectVersionedType{
			Type: typ,
		},
		Path: path,
	}
}

func (s *PathSpec) Validate(fldPath *field.Path, ctx clictx.Context, inputFilePath string) field.ErrorList {
	allErrs := field.ErrorList{}
	if s.Path == "" {
		pathField := fldPath.Child("path")
		allErrs = append(allErrs, field.Required(pathField, fmt.Sprintf("path is required for input  of type %q", s.GetType())))
	}
	return allErrs
}

type MediaFileSpec struct {
	// PathSpec holds the path that points to a file to be the base for the imput
	PathSpec `json:",inline"`

	// MediaType is the mediatype of the defined file that is also added to the oci layer.
	// Should be a custom media type in the form of "application/vnd.<mydomain>.<my description>"
	MediaType string `json:"mediaType,omitempty"`
	// CompressWithGzip defines that the blob should be automatically compressed using gzip.
	CompressWithGzip *bool `json:"compress,omitempty"`
}

func NewMediaFileSpec(typ, path, mediatype string, compress bool) MediaFileSpec {
	return MediaFileSpec{
		PathSpec:         NewPathSpec(typ, path),
		MediaType:        mediatype,
		CompressWithGzip: &compress,
	}
}

// Compress returns if the blob should be compressed using gzip.
func (s *MediaFileSpec) Compress() bool {
	if s.CompressWithGzip == nil {
		return mime.IsGZip(s.MediaType)
	}
	return *s.CompressWithGzip
}

// SetMediaTypeIfNotDefined sets the media type of the input blob if its not defined.
func (s *MediaFileSpec) SetMediaTypeIfNotDefined(mediaType string) {
	if len(s.MediaType) != 0 {
		return
	}
	s.MediaType = mediaType
}

func (s *MediaFileSpec) ValidateFile(fldPath *field.Path, ctx clictx.Context, inputFilePath string) (os.FileInfo, string, field.ErrorList) {
	allErrs := s.PathSpec.Validate(fldPath, ctx, inputFilePath)
	if s.Path != "" {
		pathField := fldPath.Child("path")
		fileInfo, filePath, err := inputs.FileInfo(ctx, s.Path, inputFilePath)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(pathField, filePath, err.Error()))
		}
		return fileInfo, filePath, allErrs
	}
	return nil, "", allErrs
}

func NewMediaFileSpecOptionType(name string, adder flagsets.ConfigAdder, types ...flagsets.ConfigOptionType) flagsets.ConfigOptionTypeSetHandler {
	set := flagsets.NewConfigOptionTypeSetHandler(name, adder, types...)
	set.AddOptionType(options.PathOption)
	set.AddOptionType(options.MediaTypeOption)
	set.AddOptionType(options.CompressOption)
	return set
}

func AddPathSpecConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	flagsets.AddFieldByOptionP(opts, options.PathOption, config, "path")
	return nil
}

func AddMediaFileSpecConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	if err := AddPathSpecConfig(opts, config); err != nil {
		return err
	}
	flagsets.AddFieldByOptionP(opts, options.MediaTypeOption, config, "mediaType")
	flagsets.AddFieldByOptionP(opts, options.CompressOption, config, "compress")
	return nil
}
