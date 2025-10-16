package cpi

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"k8s.io/apimachinery/pkg/util/validation/field"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/utils/blobaccess"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/options"
)

type PathSpec struct {
	inputs.InputSpecBase `json:",inline"`

	// Path is a file or repository path
	Path string `json:"path"`
}

func NewPathSpec(typ, path string) PathSpec {
	return PathSpec{
		InputSpecBase: inputs.InputSpecBase{
			ObjectVersionedType: runtime.ObjectVersionedType{
				Type: typ,
			},
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

////////////////////////////////////////////////////////////////////////////////

type ProcessSpec struct {
	MediaType        string `json:"mediaType,omitempty"`
	CompressWithGzip *bool  `json:"compress,omitempty"`
}

func NewProcessSpec(mediatype string, compress bool) ProcessSpec {
	return ProcessSpec{
		MediaType:        mediatype,
		CompressWithGzip: &compress,
	}
}

// Compress returns if the blob should be compressed using gzip.
func (s *ProcessSpec) Compress() bool {
	if s.CompressWithGzip == nil {
		return false
	}
	return *s.CompressWithGzip
}

// SetMediaTypeIfNotDefined sets the media type of the input blob if its not defined.
func (s *ProcessSpec) SetMediaTypeIfNotDefined(mediaType string) {
	if len(s.MediaType) != 0 {
		return
	}
	s.MediaType = mediaType
}

func (s *ProcessSpec) ProcessBlob(ctx inputs.Context, acc blobaccess.DataAccess, fs vfs.FileSystem) (blobaccess.BlobAccess, string, error) {
	if !s.Compress() {
		if s.MediaType == "" {
			s.MediaType = mime.MIME_OCTET
		}
		return blobaccess.ForDataAccess(blobaccess.BLOB_UNKNOWN_DIGEST, blobaccess.BLOB_UNKNOWN_SIZE, s.MediaType, acc), "", nil
	}

	reader, err := acc.Reader()
	if err != nil {
		return nil, "", errors.Wrapf(err, "cannot read blob data")
	}
	defer reader.Close()

	temp, err := blobaccess.NewTempFile("", "compressed*.gzip", fs)
	if err != nil {
		return nil, "", err
	}
	defer temp.Close()

	s.SetMediaTypeIfNotDefined(mime.MIME_GZIP)
	gw := gzip.NewWriter(temp.Writer())
	if _, err := io.Copy(gw, reader); err != nil {
		return nil, "", errors.Wrapf(err, "unable to compress input")
	}
	if err := gw.Close(); err != nil {
		return nil, "", errors.Wrapf(err, "unable to close gzip writer")
	}

	return temp.AsBlob(s.MediaType), "", nil
}

func AddProcessSpecOptionTypes(set flagsets.ConfigOptionTypeSetHandler) {
	set.AddOptionType(options.MediaTypeOption)
	set.AddOptionType(options.CompressOption)
}

func AddProcessSpecConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	flagsets.AddFieldByOptionP(opts, options.MediaTypeOption, config, "mediaType")
	flagsets.AddFieldByOptionP(opts, options.CompressOption, config, "compress")
	return nil
}

const ProcessSpecUsage = `
- **<code>mediaType</code>** *string*

  This OPTIONAL property describes the media type to store with the local blob.
  The default media type is ` + mime.MIME_OCTET + ` and
  ` + mime.MIME_GZIP + ` if compression is enabled.

- **<code>compress</code>** *bool*

  This OPTIONAL property describes whether the content should be stored
  compressed or not.
`

////////////////////////////////////////////////////////////////////////////////

type MediaFileSpec struct {
	// PathSpec holds the path that points to a file to be the base for the input
	PathSpec    `json:",inline"`
	ProcessSpec `json:",inline"`
}

func NewMediaFileSpec(typ, path, mediatype string, compress bool) MediaFileSpec {
	return MediaFileSpec{
		PathSpec:    NewPathSpec(typ, path),
		ProcessSpec: NewProcessSpec(mediatype, compress),
	}
}

func (s *MediaFileSpec) ValidateFile(fldPath *field.Path, ctx clictx.Context, inputFilePath string) (os.FileInfo, string, field.ErrorList) {
	allErrs := s.PathSpec.Validate(fldPath, ctx, inputFilePath)
	if s.Path != "" {
		pathField := fldPath.Child("path")
		fileInfo, filePath, err := inputs.FileInfo(ctx, s.Path, inputFilePath)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(pathField, s.Path, err.Error()))
		}
		return fileInfo, filePath, allErrs
	}
	return nil, "", allErrs
}

func NewMediaFileSpecOptionType(name string, adder flagsets.ConfigAdder, types ...flagsets.ConfigOptionType) flagsets.ConfigOptionTypeSetHandler {
	set := flagsets.NewConfigOptionTypeSetHandler(name, adder, types...)
	set.AddOptionType(options.PathOption)
	AddProcessSpecOptionTypes(set)
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
	return AddProcessSpecConfig(opts, config)
}
