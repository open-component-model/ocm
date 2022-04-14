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

package common

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	pathutil "path"
	"path/filepath"
	"strings"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/cmds/ocm/clictx"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/oci/repositories/artefactset"
	"github.com/open-component-model/ocm/pkg/oci/repositories/docker"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// MediaTypeTar defines the media type for a tarred file
const MediaTypeTar = mime.MIME_TAR

// MediaTypeGZip defines the media type for a gzipped file
const MediaTypeGZip = mime.MIME_GZIP

// MediaTypeOctetStream is the media type for any binary data.
const MediaTypeOctetStream = "application/octet-stream"

type BlobInputType string

const (
	FileInputType   BlobInputType = "file"
	DirInputType    BlobInputType = "dir"
	DockerInputType BlobInputType = "docker"
)

// BlobInput defines a local resource input that should be added to the component descriptor and
// to the resource's access.
type BlobInput struct {
	// Type defines the input type of the blob to be added.
	// Note that a input blob of type "dir" is automatically tarred.
	Type BlobInputType `json:"type"`
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
		return false
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

func (input *BlobInput) GetPath(ctx clictx.Context, inputFilePath string) (string, error) {
	fs := ctx.FileSystem()
	if input.Path == "" {
		return "", fmt.Errorf("path attribute required")
	}
	if filepath.IsAbs(input.Path) {
		return input.Path, nil
	} else {
		var wd string
		if len(inputFilePath) == 0 {
			// default to working directory if no input filepath is given
			var err error
			wd, err = fs.Getwd()
			if err != nil {
				return "", fmt.Errorf("unable to read current working directory: %w", err)
			}
		} else {
			wd = filepath.Dir(inputFilePath)
		}
		return filepath.Join(wd, input.Path), nil
	}
}

// GetBlob provides a BlobAccess for the actual input.
func (input *BlobInput) GetBlob(ctx clictx.Context, inputFilePath string) (accessio.TemporaryBlobAccess, string, error) {
	var err error
	var inputInfo os.FileInfo

	fs := ctx.FileSystem()
	inputPath := input.Path
	if input.Type != DockerInputType {
		inputPath, err = input.GetPath(ctx, inputFilePath)
		if err != nil {
			return nil, "", err
		}
		inputInfo, err = fs.Stat(inputPath)
		if err != nil {
			return nil, "", fmt.Errorf("unable to get info for input blob from %q, %w", inputPath, err)
		}
	}

	switch input.Type {
	case DockerInputType:
		locator, version, err := docker.ParseGenericRef(inputPath)
		if err != nil {
			return nil, "", err
		}
		spec := docker.NewRepositorySpec()
		repo, err := ctx.OCIContext().RepositoryForSpec(spec)
		if err != nil {
			return nil, "", err
		}
		ns, err := repo.LookupNamespace(locator)
		if err != nil {
			return nil, "", err
		}

		blob, err := artefactset.SynthesizeArtefactBlob(ns, version)
		if err != nil {
			return nil, "", err
		}
		return blob, locator, nil

	// automatically tar the input artifact if it is a directory
	case DirInputType:
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
			input.SetMediaTypeIfNotDefined(MediaTypeGZip)
			gw := gzip.NewWriter(temp.Writer())
			if err := TarFileSystem(fs, inputPath, gw, opts); err != nil {
				return nil, "", fmt.Errorf("unable to tar input artifact: %w", err)
			}
			if err := gw.Close(); err != nil {
				return nil, "", fmt.Errorf("unable to close gzip writer: %w", err)
			}
		} else {
			input.SetMediaTypeIfNotDefined(MediaTypeTar)
			if err := TarFileSystem(fs, inputPath, temp.Writer(), opts); err != nil {
				return nil, "", fmt.Errorf("unable to tar input artifact: %w", err)
			}
		}
		return temp.AsBlob(input.MediaType), "", nil

	case FileInputType:
		if inputInfo.IsDir() {
			return nil, "", fmt.Errorf("resource type is file but a directory was provided")
		}
		// otherwise just open the file
		inputBlob, err := fs.Open(inputPath)
		if err != nil {
			return nil, "", fmt.Errorf("unable to read input blob from %q: %w", inputPath, err)
		}

		if !input.Compress() {
			inputBlob.Close()
			return accessio.BlobNopCloser(accessio.BlobAccessForFile(input.MediaType, inputPath, fs)), "", nil
		}

		temp, err := accessio.NewTempFile(fs, "", "compressed*.gzip")
		if err != nil {
			return nil, "", err
		}
		defer temp.Close()

		input.SetMediaTypeIfNotDefined(MediaTypeGZip)
		gw := gzip.NewWriter(temp.Writer())
		if _, err := io.Copy(gw, inputBlob); err != nil {
			return nil, "", fmt.Errorf("unable to compress input file %q: %w", inputPath, err)
		}
		if err := gw.Close(); err != nil {
			return nil, "", fmt.Errorf("unable to close gzip writer: %w", err)
		}

		return temp.AsBlob(input.MediaType), "", nil

	default:
		return nil, "", fmt.Errorf("unknown input type %q", inputPath)
	}
}

// TarFileSystemOptions describes additional options for tarring a filesystem.
type TarFileSystemOptions struct {
	IncludeFiles []string
	ExcludeFiles []string
	// PreserveDir defines that the directory specified in the Path field should be included in the blob.
	// Only supported for Type dir.
	PreserveDir    bool
	FollowSymlinks bool

	root string
}

// Included determines whether a file should be included.
func (opts *TarFileSystemOptions) Included(path string) (bool, error) {
	// if a root path is given remove it rom the path to be checked
	if len(opts.root) != 0 {
		path = strings.TrimPrefix(path, opts.root)
	}
	// first check if a exclude regex matches
	for _, ex := range opts.ExcludeFiles {
		match, err := filepath.Match(ex, path)
		if err != nil {
			return false, fmt.Errorf("malformed filepath syntax %q", ex)
		}
		if match {
			return false, nil
		}
	}

	// if no includes are defined, include all files
	if len(opts.IncludeFiles) == 0 {
		return true, nil
	}
	// otherwise check if the file should be included
	for _, in := range opts.IncludeFiles {
		match, err := filepath.Match(in, path)
		if err != nil {
			return false, fmt.Errorf("malformed filepath syntax %q", in)
		}
		if match {
			return true, nil
		}
	}
	return false, nil
}

// TarFileSystem creates a tar archive from a filesystem.
func TarFileSystem(fs vfs.FileSystem, root string, writer io.Writer, opts TarFileSystemOptions) error {
	tw := tar.NewWriter(writer)
	if opts.PreserveDir {
		opts.root = pathutil.Base(root)
	}
	if err := addFileToTar(fs, tw, opts.root, root, opts); err != nil {
		return err
	}
	return tw.Close()
}

func addFileToTar(fs vfs.FileSystem, tw *tar.Writer, path string, realPath string, opts TarFileSystemOptions) error {
	if len(path) != 0 { // do not check the root
		include, err := opts.Included(path)
		if err != nil {
			return err
		}
		if !include {
			return nil
		}
	}
	info, err := fs.Lstat(realPath)
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = path

	switch {
	case info.IsDir():
		// do not write root header
		if len(path) != 0 {
			if err := tw.WriteHeader(header); err != nil {
				return fmt.Errorf("unable to write header for %q: %w", path, err)
			}
		}
		err := vfs.Walk(fs, realPath, func(subFilePath string, info os.FileInfo, err error) error {
			if subFilePath == realPath {
				return nil
			}
			if err != nil {
				return err
			}
			relPath, err := filepath.Rel(realPath, subFilePath)
			if err != nil {
				return fmt.Errorf("unable to calculate relative path for %s: %w", subFilePath, err)
			}
			return addFileToTar(fs, tw, pathutil.Join(path, relPath), subFilePath, opts)
		})
		return err
	case info.Mode().IsRegular():
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("unable to write header for %q: %w", path, err)
		}
		file, err := fs.OpenFile(realPath, os.O_RDONLY, os.ModePerm)
		if err != nil {
			return fmt.Errorf("unable to open file %q: %w", path, err)
		}
		if _, err := io.Copy(tw, file); err != nil {
			_ = file.Close()
			return fmt.Errorf("unable to add file to tar %q: %w", path, err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("unable to close file %q: %w", path, err)
		}
		return nil
	case header.Typeflag == tar.TypeSymlink:
		if !opts.FollowSymlinks {
			//log.Info(fmt.Sprintf("symlink found in %q but symlinks are not followed", path))
			return nil
		}
		realPath, err := vfs.EvalSymlinks(fs, realPath)
		if err != nil {
			return fmt.Errorf("unable to follow symlink %s: %w", realPath, err)
		}
		return addFileToTar(fs, tw, path, realPath, opts)
	default:
		return fmt.Errorf("unsupported file type %s in %s", info.Mode().String(), path)
	}
}

////////////////////////////////////////////////////////////////////////////////

func (input *BlobInput) Validate(fldPath *field.Path, ctx clictx.Context, inputFilePath string) field.ErrorList {
	if input == nil {
		return nil
	}
	allErrs := field.ErrorList{}
	if input.Type != DirInputType && input.Type != FileInputType && input.Type != DockerInputType {
		path := fldPath.Child("type")
		if input.Type == "" {
			allErrs = append(allErrs, field.Required(path, "input type required"))
		} else {
			allErrs = append(allErrs, field.NotSupported(path, input.Type, []string{string(DirInputType), string(FileInputType)}))
		}
	} else {
		pathField := fldPath.Child("path")
		if input.Path == "" {
			allErrs = append(allErrs, field.Required(pathField, "path is required for input"))
		} else {
			if input.Type != DockerInputType {
				fs := ctx.FileSystem()
				filePath, err := input.GetPath(ctx, inputFilePath)
				if err != nil {
					allErrs = append(allErrs, field.Invalid(pathField, filePath, err.Error()))
				} else {
					ok, err := vfs.Exists(fs, filePath)
					if err != nil {
						allErrs = append(allErrs, field.Invalid(pathField, filePath, err.Error()))
					} else {
						if !ok {
							allErrs = append(allErrs, field.NotFound(pathField, filePath))
						}
					}

					if input.Type == DirInputType {
						if ok {
							ok, err := vfs.DirExists(fs, filePath)
							if err != nil {
								allErrs = append(allErrs, field.Invalid(pathField, filePath, err.Error()))
							} else {
								if !ok {
									allErrs = append(allErrs, field.Invalid(pathField, filePath, "no directory"))
								}
							}
						}
					} else {
						if ok {
							ok, err := vfs.FileExists(fs, filePath)
							if err != nil {
								allErrs = append(allErrs, field.Invalid(pathField, filePath, err.Error()))
							} else {
								if !ok {
									allErrs = append(allErrs, field.Invalid(pathField, filePath, "no regular file"))
								}
							}
						}
						if input.PreserveDir != nil {
							allErrs = append(allErrs, field.Forbidden(fldPath.Child("preserveDir"), "only supported for type "+string(DirInputType)))
						}
						if input.FollowSymlinks != nil {
							allErrs = append(allErrs, field.Forbidden(fldPath.Child("followSymlinks"), "only supported for type "+string(DirInputType)))
						}
						if input.CompressWithGzip != nil {
							allErrs = append(allErrs, field.Forbidden(fldPath.Child("compress"), "only supported for type "+string(DirInputType)))
						}
						if input.ExcludeFiles != nil {
							allErrs = append(allErrs, field.Forbidden(fldPath.Child("excludeFiles"), "only supported for type "+string(DirInputType)))
						}
						if input.IncludeFiles != nil {
							allErrs = append(allErrs, field.Forbidden(fldPath.Child("includeFiles"), "only supported for type "+string(DirInputType)))
						}
					}
				}
			} else {
				_, _, err := docker.ParseGenericRef(input.Path)
				if err != nil {
					allErrs = append(allErrs, field.Invalid(pathField, input.Path, err.Error()))

				}
			}
		}
	}
	return allErrs
}
