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
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gardener/ocm/pkg/ocm/compdesc"
	metav1 "github.com/gardener/ocm/pkg/ocm/compdesc/meta/v1"
	"github.com/gardener/ocm/pkg/ocm/core"
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"

	"github.com/gardener/component-spec/bindings-go/codec"
)

// ComponentDescriptorFileName is the name of the component-descriptor file.
const ComponentDescriptorFileName = "component-descriptor.yaml"

// BlobsDirectoryName is the name of the blob directory in the tar.
const BlobsDirectoryName = "blobs"

var UnsupportedResolveType = errors.New("UnsupportedResolveType")

// ComponentArchive is the go representation for a component artefact
type ComponentArchive struct {
	ComponentDescriptor *compdesc.ComponentDescriptor
	fs                  vfs.FileSystem
	ctx                 core.Context
}

var _ core.ComponentAccess = &ComponentArchive{}

// NewComponentArchive returns a new component descriptor with a filesystem
func NewComponentArchive(ctx core.Context, cd *compdesc.ComponentDescriptor, fs vfs.FileSystem) *ComponentArchive {
	return &ComponentArchive{
		ComponentDescriptor: cd,
		fs:                  fs,
		ctx:                 ctx,
	}
}
func (c *ComponentArchive) GetContext() core.Context {
	return c.ctx
}

func (c *ComponentArchive) GetAccessType() string {
	return CTFRepositoryType
}

func (c *ComponentArchive) Close() error {
	// TODO
	return vfs.Cleanup(c.fs)
}

func (c *ComponentArchive) GetRepository() core.Repository {
	// TODO
	return nil
}

func (c *ComponentArchive) GetName() string {
	return c.ComponentDescriptor.GetName()
}

func (c *ComponentArchive) GetVersion() string {
	return c.ComponentDescriptor.GetVersion()
}

func (c *ComponentArchive) GetDescriptor() (*compdesc.ComponentDescriptor, error) {
	return c.ComponentDescriptor, nil
}

func (c *ComponentArchive) GetResource(meta *metav1.Identity) (core.ResourceAccess, error) {
	// TODO
	return nil, fmt.Errorf("not implemented")
}
func (c *ComponentArchive) GetSource(meta *metav1.Identity) (core.ResourceAccess, error) {
	// TODO
	return nil, fmt.Errorf("not implemented")
}

var _osfs = osfs.New()

func FileSystem(ofs []vfs.FileSystem) vfs.FileSystem {
	if len(ofs) > 0 {
		return ofs[0]
	}
	return _osfs
}

// OpenComponentArchiveFromDirectory creates a component archive from a path
func OpenComponentArchiveFromDirectory(ctx core.Context, path string, ofs ...vfs.FileSystem) (*ComponentArchive, error) {
	fs := FileSystem(ofs)

	fs, err := projectionfs.New(fs, path)
	if err != nil {
		return nil, fmt.Errorf("unable to create projected filesystem from path %s: %w", path, err)
	}

	return OpenComponentArchiveFromFilesystem(ctx, fs)
}

// OpenComponentArchive creates a new component archive from a file path.
func OpenComponentArchive(ctx core.Context, path string, ofs ...vfs.FileSystem) (*ComponentArchive, error) {
	fs := FileSystem(ofs)

	fi, err := fs.Stat(path)
	if err != nil {
		return nil, err
	}
	if fi.IsDir() {
		return OpenComponentArchiveFromDirectory(ctx, path, fs)
	}

	ca, err := OpenComponentArchiveFromTGZ(ctx, path, fs)
	if err != nil {
		ca, err = OpenComponentArchiveFromTAR(ctx, path, fs)
	}
	return ca, err
}

func OpenComponentArchiveFromTGZ(ctx core.Context, path string, ofs ...vfs.FileSystem) (*ComponentArchive, error) {
	fs := FileSystem(ofs)
	// we expect that the path point to a tar or tgz
	file, err := fs.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open tar archive from %s: %w", path, err)
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("unable to open gzip reader for %s: %w", path, err)
	}
	return OpenComponentArchiveFromTarReader(ctx, reader)
}

// OpenComponentArchiveFromTAR creates a new componet archive from a component tar file.
func OpenComponentArchiveFromTAR(ctx core.Context, path string, ofs ...vfs.FileSystem) (*ComponentArchive, error) {
	fs := FileSystem(ofs)
	// we expect that the path point to a tar
	file, err := fs.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open tar archive from %s: %w", path, err)
	}
	defer file.Close()
	return OpenComponentArchiveFromTarReader(ctx, file)
}

// OpenComponentArchiveFromTarReader creates a new manifest builder from a input reader.
func OpenComponentArchiveFromTarReader(ctx core.Context, in io.Reader, tfs ...vfs.FileSystem) (*ComponentArchive, error) {
	// the archive is untared to a memory fs that the builder can work
	// as it would be a default filesystem.
	var fs vfs.FileSystem
	if len(tfs) == 0 {
		fs = memoryfs.New()
	} else {
		fs = tfs[0]
	}

	if err := ExtractTarToFs(fs, in); err != nil {
		return nil, fmt.Errorf("unable to extract tar: %w", err)
	}

	return OpenComponentArchiveFromFilesystem(ctx, fs)
}

// NewComponentArchiveFromFilesystem creates a new component archive from a filesystem.
func OpenComponentArchiveFromFilesystem(ctx core.Context, fs vfs.FileSystem, decodeOpts ...compdesc.DecodeOption) (*ComponentArchive, error) {
	data, err := vfs.ReadFile(fs, filepath.Join("/", ComponentDescriptorFileName))
	if err != nil {
		return nil, fmt.Errorf("unable to read the component descriptor from %s: %w", ComponentDescriptorFileName, err)
	}
	cd, err := compdesc.Decode(data, decodeOpts...)
	if err != nil {
		return nil, fmt.Errorf("unable to parse component descriptor read from %s: %w", ComponentDescriptorFileName, err)
	}

	return NewComponentArchive(ctx, cd, fs), nil
}

// Digest returns the digest of the component archive.
// The digest is computed serializing the included component descriptor into json and compute sha hash.
func (ca *ComponentArchive) Digest() (string, error) {
	data, err := codec.Encode(ca.ComponentDescriptor)
	if err != nil {
		return "", err
	}
	return digest.FromBytes(data).String(), nil
}

// AddSource adds a blob source to the current archive.
// If the specified source already exists it will be overwritten.
func (ca *ComponentArchive) AddSource(meta *core.SourceMeta, acc core.BlobAccess) error {
	if acc == nil {
		return errors.New("a source has to be defined")
	}
	id := ca.ComponentDescriptor.GetSourceIndex(meta)
	if err := ca.ensureBlobsPath(); err != nil {
		return err
	}

	digest, err := core.Digest(acc)
	if err != nil {
		return err
	}

	blobpath := BlobPath(string(digest))
	if _, err := ca.fs.Stat(blobpath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("unable to get file info for %s", blobpath)
		}
		file, err := ca.fs.OpenFile(blobpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return fmt.Errorf("unable to open file %s: %w", blobpath, err)
		}
		src, err := acc.Reader()
		if err != nil {
			return err
		}
		defer src.Close()
		if _, err := io.Copy(file, src); err != nil {
			return fmt.Errorf("unable to write blob %q to file: %w", blobpath, err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("unable to close file: %w", err)
		}
	}

	src := &compdesc.Source{
		SourceMeta: *meta.Copy(),
		Access:     NewLocalFilesystemBlobAccessSpecV1(string(digest), acc.MimeType()),
	}

	if id == -1 {
		ca.ComponentDescriptor.Sources = append(ca.ComponentDescriptor.Sources, *src)
	} else {
		ca.ComponentDescriptor.Sources[id] = *src
	}
	return nil
}

// AddResource adds a blob resource to the current archive.
func (ca *ComponentArchive) AddResource(meta *core.ResourceMeta, acc core.BlobAccess) error {
	if acc == nil {
		return errors.New("a resource has to be defined")
	}
	idx := ca.ComponentDescriptor.GetResourceIndex(meta)
	if err := ca.ensureBlobsPath(); err != nil {
		return err
	}

	digest, err := core.Digest(acc)
	if err != nil {
		return err
	}

	blobpath := BlobPath(string(digest))
	if _, err := ca.fs.Stat(blobpath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("unable to get file info for %s", blobpath)
		}
		file, err := ca.fs.OpenFile(blobpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return fmt.Errorf("unable to open file %s: %w", blobpath, err)
		}
		src, err := acc.Reader()
		if err != nil {
			return err
		}
		defer src.Close()

		_, err = io.Copy(file, src)
		if err != nil {
			return fmt.Errorf("unable to write blob to file %q: %w", blobpath, err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("unable to close file: %w", err)
		}
		return nil
	}

	res := &compdesc.Resource{
		ResourceMeta: *meta.Copy(),
		Access:       NewLocalFilesystemBlobAccessSpecV1(string(digest), acc.MimeType()),
	}

	if idx == -1 {
		ca.ComponentDescriptor.Resources = append(ca.ComponentDescriptor.Resources, *res)
	} else {
		ca.ComponentDescriptor.Resources[idx] = *res
	}
	return nil
}

// ensureBlobsPath ensures that the blob directory exists
func (ca *ComponentArchive) ensureBlobsPath() error {
	if _, err := ca.fs.Stat(BlobsDirectoryName); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("unable to get file info for blob directory: %w", err)
		}
		return ca.fs.Mkdir(BlobsDirectoryName, os.ModePerm)
	}
	return nil
}

// WriteTarGzip tars the current components descriptor and its artifacts.
func (ca *ComponentArchive) WriteTarGzip(writer io.Writer) error {
	gw := gzip.NewWriter(writer)
	if err := ca.WriteTar(gw); err != nil {
		return err
	}
	return gw.Close()
}

// WriteTar tars the current components descriptor and its artifacts.
func (ca *ComponentArchive) WriteTar(writer io.Writer) error {
	tw := tar.NewWriter(writer)

	// write component descriptor
	cdBytes, err := codec.Encode(ca.ComponentDescriptor)
	if err != nil {
		return fmt.Errorf("unable to encode component descriptor: %w", err)
	}
	cdHeader := &tar.Header{
		Name:    ComponentDescriptorFileName,
		Size:    int64(len(cdBytes)),
		Mode:    0644,
		ModTime: time.Now(),
	}

	if err := tw.WriteHeader(cdHeader); err != nil {
		return fmt.Errorf("unable to write component descriptor header: %w", err)
	}
	if _, err := io.Copy(tw, bytes.NewBuffer(cdBytes)); err != nil {
		return fmt.Errorf("unable to write component descriptor content: %w", err)
	}

	// add all blobs
	err = tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeDir,
		Name:     BlobsDirectoryName,
		Mode:     0644,
		ModTime:  time.Now(),
	})
	if err != nil {
		return fmt.Errorf("unable to write blob directory: %w", err)
	}

	blobs, err := vfs.ReadDir(ca.fs, BlobsDirectoryName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("unable to read blob directory: %w", err)
	}
	for _, blobInfo := range blobs {
		blobpath := BlobPath(blobInfo.Name())
		header := &tar.Header{
			Name:    blobpath,
			Size:    blobInfo.Size(),
			Mode:    0644,
			ModTime: time.Now(),
		}
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("unable to write blob header: %w", err)
		}

		blob, err := ca.fs.Open(blobpath)
		if err != nil {
			return fmt.Errorf("unable to open blob: %w", err)
		}
		if _, err := io.Copy(tw, blob); err != nil {
			return fmt.Errorf("unable to write blob content: %w", err)
		}
		if err := blob.Close(); err != nil {
			return fmt.Errorf("unable to close blob %s: %w", blobpath, err)
		}
	}

	return tw.Close()
}

// WriteToFilesystem writes the current component archive to a filesystem
func (ca *ComponentArchive) WriteToFilesystem(fs vfs.FileSystem, path string) error {
	// create the directory structure with the blob directory
	if err := fs.MkdirAll(filepath.Join(path, BlobsDirectoryName), os.ModePerm); err != nil {
		return fmt.Errorf("unable to create output directory %q: %s", path, err.Error())
	}
	// copy component-descriptor
	cdBytes, err := codec.Encode(ca.ComponentDescriptor)
	if err != nil {
		return fmt.Errorf("unable to encode component descriptor: %w", err)
	}
	if err := vfs.WriteFile(fs, filepath.Join(path, ComponentDescriptorFileName), cdBytes, os.ModePerm); err != nil {
		return fmt.Errorf("unable to copy component descritptor to %q: %w", filepath.Join(path, ComponentDescriptorFileName), err)
	}

	// copy all blobs
	blobInfos, err := vfs.ReadDir(ca.fs, BlobsDirectoryName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("unable to read blobs: %w", err)
	}
	for _, blobInfo := range blobInfos {
		if blobInfo.IsDir() {
			continue
		}
		inpath := BlobPath(blobInfo.Name())
		outpath := filepath.Join(path, BlobsDirectoryName, blobInfo.Name())
		blob, err := ca.fs.Open(inpath)
		if err != nil {
			return fmt.Errorf("unable to open input blob %q: %w", inpath, err)
		}
		out, err := fs.OpenFile(outpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return fmt.Errorf("unable to open output blob %q: %w", outpath, err)
		}
		if _, err := io.Copy(out, blob); err != nil {
			return fmt.Errorf("unable to copy blob from %q to %q: %w", inpath, outpath, err)
		}
		if err := out.Close(); err != nil {
			return fmt.Errorf("unable to close output blob %s: %w", outpath, err)
		}
		if err := blob.Close(); err != nil {
			return fmt.Errorf("unable to close input blob %s: %w", outpath, err)
		}
	}

	return nil
}

// BlobPath returns the path to the blob for a given name.
func BlobPath(name string) string {
	return filepath.Join(BlobsDirectoryName, name)
}

// ExtractTarToFs writes a tar stream to a filesystem.
func ExtractTarToFs(fs vfs.FileSystem, in io.Reader) error {
	tr := tar.NewReader(in)
	for {
		header, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := fs.MkdirAll(header.Name, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("unable to create directory %s: %w", header.Name, err)
			}
		case tar.TypeReg:
			file, err := fs.OpenFile(header.Name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("unable to open file %s: %w", header.Name, err)
			}
			if _, err := io.Copy(file, tr); err != nil {
				return fmt.Errorf("unable to copy tar file to filesystem: %w", err)
			}
			if err := file.Close(); err != nil {
				return fmt.Errorf("unable to close file %s: %w", header.Name, err)
			}
		}
	}
}
