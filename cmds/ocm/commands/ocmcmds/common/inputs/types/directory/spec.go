package directory

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/blobaccess"
	"ocm.software/ocm/api/utils/blobaccess/dirtree"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
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

func (s *Spec) GetBlob(ctx inputs.Context, info inputs.InputResourceInfo) (blobaccess.BlobAccess, string, error) {
	fs := ctx.FileSystem()
	inputInfo, inputPath, err := inputs.FileInfo(ctx, s.Path, info.InputFilePath)
	if err != nil {
		return nil, "", fmt.Errorf("resource dir %s: %w", info.InputFilePath, err)
	}
	if !inputInfo.IsDir() {
		return nil, "", fmt.Errorf("resource type is dir but a file was provided")
	}

	access, err := dirtree.BlobAccess(inputPath,
		dirtree.WithMimeType(s.MediaType),
		dirtree.WithFileSystem(fs),
		dirtree.WithCompressWithGzip(s.Compress()),
		dirtree.WithIncludeFiles(s.IncludeFiles),
		dirtree.WithExcludeFiles(s.ExcludeFiles),
		dirtree.WithFollowSymlinks(utils.AsBool(s.FollowSymlinks)),
		dirtree.WithPreserveDir(utils.AsBool(s.PreserveDir)),
	)
	return access, "", err
}
