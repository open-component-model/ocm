package accessobj

import (
	"archive/tar"
	"fmt"
	"io"
	"os"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/compression"
	"ocm.software/ocm/api/utils/tarutils"
)

var FormatTAR = NewTarHandler()

func init() {
	RegisterFormat(FormatTAR)
}

type TarHandler struct {
	format      FileFormat
	compression compression.Algorithm
}

var _ StandardReaderHandler = (*TarHandler)(nil)

var _ FormatHandler = (*TarHandler)(nil)

func NewTarHandler() *TarHandler {
	return NewTarHandlerWithCompression(accessio.FormatTar, nil)
}

func NewTarHandlerWithCompression(format FileFormat, algorithm compression.Algorithm) *TarHandler {
	return &TarHandler{
		format:      format,
		compression: algorithm,
	}
}

// ApplyOption applies the configured path filesystem.
func (h *TarHandler) ApplyOption(options accessio.Options) error {
	options.SetFileFormat(h.Format())
	return nil
}

func (h *TarHandler) Format() accessio.FileFormat {
	return h.format
}

func (h *TarHandler) Open(info AccessObjectInfo, acc AccessMode, path string, opts accessio.Options) (*AccessObject, error) {
	return DefaultOpenOptsFileHandling(fmt.Sprintf("%s archive", h.format), info, acc, path, opts, h)
}

func (h *TarHandler) Create(info AccessObjectInfo, path string, opts accessio.Options, mode vfs.FileMode) (*AccessObject, error) {
	return DefaultCreateOptsFileHandling(fmt.Sprintf("%s archive", h.format), info, path, opts, mode, h)
}

// Write tars the current descriptor and its artifacts.
func (h *TarHandler) Write(obj *AccessObject, path string, opts accessio.Options, mode vfs.FileMode) (err error) {
	writer, err := opts.WriterFor(path, mode)
	if err != nil {
		return fmt.Errorf("unable to write: %w", err)
	}

	defer func() {
		err = errors.Join(err, writer.Close())
	}()

	// Check if OCI layout (has subdirectories in element dir)
	if h.hasNestedDirs(obj) {
		return h.writeToStream(obj, writer, opts, h.writeOCICompliant)
	}
	return h.writeToStream(obj, writer, opts, h.writeElementsFlat)
}

// hasNestedDirs checks if the element directory contains subdirectories (OCI layout).
func (h *TarHandler) hasNestedDirs(obj *AccessObject) bool {
	elemDir := obj.info.GetElementDirectoryName()
	entries, err := vfs.ReadDir(obj.fs, elemDir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() {
			return true
		}
	}
	return false
}

// writeToStream is the common implementation for writing tar streams.
// The elementWriter parameter determines how element content is written (flat vs nested).
func (h TarHandler) writeToStream(obj *AccessObject, writer io.Writer, opts accessio.Options, elementWriter func(*AccessObject, *tar.Writer) error) error {
	writer, cleanup, err := h.applyCompression(writer)
	if err != nil {
		return err
	}
	if cleanup != nil {
		defer cleanup()
	}

	tw := tar.NewWriter(writer)

	if err := h.writeDescriptor(obj, tw); err != nil {
		return err
	}

	if err := h.writeAdditionalFiles(obj, tw); err != nil {
		return err
	}

	if err := elementWriter(obj, tw); err != nil {
		return err
	}

	return tw.Close()
}

// applyCompression wraps the writer with compression if configured.
// Returns the wrapped writer and a cleanup function (may be nil).
func (h TarHandler) applyCompression(writer io.Writer) (io.Writer, func(), error) {
	if h.compression == nil {
		return writer, nil, nil
	}
	w, err := h.compression.Compressor(writer, nil, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to compress writer: %w", err)
	}
	return w, func() { w.Close() }, nil
}

// writeDescriptor updates the access object and writes the descriptor to the tar.
func (h TarHandler) writeDescriptor(obj *AccessObject, tw *tar.Writer) error {
	if _, err := obj.Update(); err != nil {
		return fmt.Errorf("unable to update access object: %w", err)
	}

	data, err := obj.state.GetBlob()
	if err != nil {
		return fmt.Errorf("unable to get state blob: %w", err)
	}
	defer data.Close()

	header := &tar.Header{
		Name:    obj.info.GetDescriptorFileName(),
		Size:    data.Size(),
		Mode:    FileMode,
		ModTime: ModTime,
	}
	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("unable to write descriptor header: %w", err)
	}

	r, err := data.Reader()
	if err != nil {
		return fmt.Errorf("unable to get reader: %w", err)
	}
	defer r.Close()

	if _, err := io.CopyN(tw, r, data.Size()); err != nil {
		return fmt.Errorf("unable to write descriptor content: %w", err)
	}
	return nil
}

// writeAdditionalFiles copies additional files to the tar.
func (h TarHandler) writeAdditionalFiles(obj *AccessObject, tw *tar.Writer) error {
	for _, f := range obj.info.GetAdditionalFiles(obj.fs) {
		if err := h.writeFileIfExists(obj, tw, f); err != nil {
			return err
		}
	}
	return nil
}

// writeFileIfExists writes a single file to the tar if it exists.
func (h TarHandler) writeFileIfExists(obj *AccessObject, tw *tar.Writer, path string) (err error) {
	ok, err := vfs.IsFile(obj.fs, path)
	if err != nil {
		return errors.Wrapf(err, "cannot check for file %q", path)
	}
	if !ok {
		return nil
	}

	fi, err := obj.fs.Stat(path)
	if err != nil {
		return errors.Wrapf(err, "cannot stat file %q", path)
	}

	header := &tar.Header{
		Name:    path,
		Size:    fi.Size(),
		Mode:    FileMode,
		ModTime: ModTime,
	}
	if err := tw.WriteHeader(header); err != nil {
		return errors.Wrapf(err, "unable to write header for %q", path)
	}

	r, err := obj.fs.Open(path)
	if err != nil {
		return errors.Wrapf(err, "unable to open %q", path)
	}
	defer func() {
		err = errors.Join(err, r.Close())
	}()

	if _, err := io.CopyN(tw, r, fi.Size()); err != nil {
		return errors.Wrapf(err, "unable to write file %q", path)
	}
	return nil
}

// writeElementsFlat writes element content using flat directory structure.
func (h TarHandler) writeElementsFlat(obj *AccessObject, tw *tar.Writer) error {
	elemDir := obj.info.GetElementDirectoryName()

	if err := tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeDir,
		Name:     elemDir,
		Mode:     DirMode,
		ModTime:  ModTime,
	}); err != nil {
		return fmt.Errorf("unable to write %s directory: %w", obj.info.GetElementTypeName(), err)
	}

	fileInfos, err := vfs.ReadDir(obj.fs, elemDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("unable to read %s directory: %w", obj.info.GetElementTypeName(), err)
	}

	for _, fileInfo := range fileInfos {
		path := obj.info.SubPath(fileInfo.Name())
		if err := h.writeFileEntry(obj, tw, path, fileInfo); err != nil {
			return err
		}
	}
	return nil
}

// writeFileEntry writes a single file entry to the tar.
func (h TarHandler) writeFileEntry(obj *AccessObject, tw *tar.Writer, path string, info os.FileInfo) (err error) {
	header := &tar.Header{
		Name:    path,
		Size:    info.Size(),
		Mode:    FileMode,
		ModTime: ModTime,
	}
	if err = tw.WriteHeader(header); err != nil {
		return fmt.Errorf("unable to write %s header: %w", obj.info.GetElementTypeName(), err)
	}

	content, err := obj.fs.Open(path)
	if err != nil {
		return fmt.Errorf("unable to open %s: %w", obj.info.GetElementTypeName(), err)
	}
	defer func() {
		err = errors.Join(err, content.Close())
	}()

	if _, err = io.CopyN(tw, content, info.Size()); err != nil {
		return fmt.Errorf("unable to write %s content: %w", obj.info.GetElementTypeName(), err)
	}
	return nil
}

// writeOCICompliant writes element content to the tar following OCI standards.
// This handles OCI layouts with a standard two-level structure (e.g. blobs/sha256).
// See: https://specs.opencontainers.org/image-spec/image-layout/?v=v1.1.1#filesystem-layout
func (h TarHandler) writeOCICompliant(obj *AccessObject, tw *tar.Writer) error {
	dir := obj.info.GetElementDirectoryName()
	// Write root directory header
	if err := tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeDir,
		Name:     dir,
		Mode:     DirMode,
		ModTime:  ModTime,
	}); err != nil {
		return fmt.Errorf("unable to write directory header for %s: %w", dir, err)
	}

	entries, err := vfs.ReadDir(obj.fs, dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("unable to read directory %s: %w", dir, err)
	}

	// Process first level entries (typically 'blobs')
	for _, entry := range entries {
		subPath := dir + "/" + entry.Name()

		if entry.IsDir() {
			if err := h.writeDirEntry(obj, tw, subPath); err != nil {
				return err
			}
		} else {
			// Handle files in root directory
			if err := h.writeFileEntry(obj, tw, subPath, entry); err != nil {
				return err
			}
		}
	}
	return nil
}

// writeDirEntry writes a directory entry and its content to the tar.
// This specifically handles OCI-style directory entries (e.g. the 'sha256' subdirectory).
func (h TarHandler) writeDirEntry(obj *AccessObject, tw *tar.Writer, path string) error {
	// Write directory header
	if err := tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeDir,
		Name:     path,
		Mode:     DirMode,
		ModTime:  ModTime,
	}); err != nil {
		return fmt.Errorf("unable to write directory header for %s: %w", path, err)
	}

	// Process entries in this directory (typically hash digest files)
	entries, err := vfs.ReadDir(obj.fs, path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("unable to read directory %s: %w", path, err)
	}

	// Process files in directory
	for _, entry := range entries {
		filePath := path + "/" + entry.Name()
		if !entry.IsDir() {
			if err := h.writeFileEntry(obj, tw, filePath, entry); err != nil {
				return err
			}
		}
	}

	return nil
}

func (h *TarHandler) NewFromReader(info AccessObjectInfo, acc AccessMode, in io.Reader, opts accessio.Options, closer Closer) (*AccessObject, error) {
	if h.compression != nil {
		reader, err := h.compression.Decompressor(in)
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		in = reader
	}
	setup := func(fs vfs.FileSystem) error {
		if err := tarutils.ExtractTarToFs(fs, in); err != nil {
			return fmt.Errorf("unable to extract tar: %w", err)
		}
		return nil
	}
	return NewAccessObject(info, acc, opts.GetRepresentation(), SetupFunction(setup), closer, DirMode)
}
