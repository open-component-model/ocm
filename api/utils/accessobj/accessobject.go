package accessobj

import (
	"archive/tar"
	"fmt"
	"io"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/compression"
)

type DescriptorHandlerFactory func(fs vfs.FileSystem) StateHandler

////////////////////////////////////////////////////////////////////////////////

// AccessObjectInfo is used to control the persistence of
// a serialization format for sets of elements.
type AccessObjectInfo interface {
	SetupFor(fs vfs.FileSystem) error
	GetDescriptorFileName() string
	GetObjectTypeName() string
	GetElementTypeName() string
	GetElementDirectoryName() string
	GetAdditionalFiles(fs vfs.FileSystem) []string
	SetupFileSystem(fs vfs.FileSystem, mode vfs.FileMode) error
	SetupDescriptorState(fs vfs.FileSystem) StateHandler
	SubPath(name string) string

	ValidateDescriptor(data []byte) error
}

// DefaultAccessObjectInfo is a default implementation for AccessObjectInfo
// that can be used to describe a simple static configuration.
// The methods do not change the content, therefore an instance can be reused.
type DefaultAccessObjectInfo struct {
	DescriptorFileName       string
	ObjectTypeName           string
	ElementDirectoryName     string
	ElementTypeName          string
	DescriptorHandlerFactory DescriptorHandlerFactory
	AdditionalFiles          []string
	DescriptorValidator      func([]byte) error
}

var _ AccessObjectInfo = (*DefaultAccessObjectInfo)(nil)

func (i *DefaultAccessObjectInfo) SetupFor(fs vfs.FileSystem) error {
	return nil
}

func (i *DefaultAccessObjectInfo) GetDescriptorFileName() string {
	return i.DescriptorFileName
}

func (i *DefaultAccessObjectInfo) GetObjectTypeName() string {
	return i.ObjectTypeName
}

func (i *DefaultAccessObjectInfo) GetElementTypeName() string {
	return i.ElementTypeName
}

func (i *DefaultAccessObjectInfo) GetElementDirectoryName() string {
	return i.ElementDirectoryName
}

func (i *DefaultAccessObjectInfo) GetAdditionalFiles(fs vfs.FileSystem) []string {
	return i.AdditionalFiles
}

func (i *DefaultAccessObjectInfo) SetupFileSystem(fs vfs.FileSystem, mode vfs.FileMode) error {
	if i.ElementDirectoryName != "" {
		return fs.MkdirAll(i.ElementDirectoryName, mode)
	}
	return nil
}

func (i *DefaultAccessObjectInfo) SetupDescriptorState(fs vfs.FileSystem) StateHandler {
	return i.DescriptorHandlerFactory(fs)
}

func (i *DefaultAccessObjectInfo) SubPath(name string) string {
	return filepath.Join(i.ElementDirectoryName, name)
}

func (i *DefaultAccessObjectInfo) ValidateDescriptor(data []byte) error {
	if i.DescriptorValidator != nil {
		return i.DescriptorValidator(data)
	}
	return nil
}

// AccessObject provides a basic functionality for descriptor based access objects
// using a virtual filesystem for the internal representation.
type AccessObject struct {
	info    AccessObjectInfo
	fs      vfs.FileSystem
	cleanup bool
	mode    vfs.FileMode
	state   State
	closer  Closer
}

func NewAccessObject(info AccessObjectInfo, acc AccessMode, fs vfs.FileSystem, setup Setup, closer Closer, mode vfs.FileMode) (*AccessObject, error) {
	defaulted, fs, err := InternalRepresentationFilesystem(acc, fs, info.SetupFileSystem, mode)
	if err != nil {
		return nil, err
	}
	if setup != nil {
		if err := setup.Setup(fs); err != nil {
			return nil, err
		}
	}
	if err := info.SetupFor(fs); err != nil {
		return nil, err
	}

	s, err := NewFileBasedState(acc, fs, info.GetDescriptorFileName(), "", info.SetupDescriptorState(fs), mode)
	if err != nil {
		return nil, err
	}
	obj := &AccessObject{
		info:    info,
		state:   s,
		fs:      fs,
		cleanup: defaulted,
		mode:    mode,
		closer:  closer,
	}

	return obj, nil
}

func (a *AccessObject) GetInfo() AccessObjectInfo {
	return a.info
}

func (a *AccessObject) GetFileSystem() vfs.FileSystem {
	return a.fs
}

func (a *AccessObject) GetMode() vfs.FileMode {
	return a.mode
}

func (a *AccessObject) GetState() State {
	return a.state
}

func (a *AccessObject) IsClosed() bool {
	return a.fs == nil
}

func (a *AccessObject) IsReadOnly() bool {
	return a.state.IsReadOnly()
}

func (a *AccessObject) SetReadOnly() {
	a.state.SetReadOnly()
}

func (a *AccessObject) updateDescriptor() (bool, error) {
	if a.IsClosed() {
		return false, accessio.ErrClosed
	}
	return a.state.Update()
}

func (a *AccessObject) Write(path string, mode vfs.FileMode, opts ...accessio.Option) error {
	if a.IsClosed() {
		return accessio.ErrClosed
	}

	o, err := accessio.AccessOptions(nil, opts...)
	if err != nil {
		return err
	}

	f := GetFormat(*o.GetFileFormat())
	if f == nil {
		return errors.ErrUnknown("file format", string(*o.GetFileFormat()))
	}

	return f.Write(a, path, o, mode)
}

func (a *AccessObject) Update() (bool, error) {
	if b, err := a.updateDescriptor(); err != nil {
		return b, fmt.Errorf("unable to update descriptor: %w", err)
	} else {
		return b, nil
	}
}

func (a *AccessObject) Close() error {
	if a.IsClosed() {
		return accessio.ErrClosed
	}
	list := errors.ErrListf("cannot close %s", a.info.GetObjectTypeName())
	_, err := a.Update()
	list.Add(err)
	if a.closer != nil {
		list.Add(a.closer.Close(a))
	}
	if a.cleanup {
		list.Add(vfs.Cleanup(a.fs))
	}
	a.fs = nil
	return list.Result()
}

////////////////////////////////////////////////////////////////////////////////

func ValidateDescriptor(info AccessObjectInfo, path string, fss ...vfs.FileSystem) error {
	fs := utils.FileSystem(fss...)
	_, err := vfs.Canonical(fs, path, true)
	if err != nil {
		return err
	}
	ok, err := vfs.Exists(fs, path)
	if err != nil {
		return err
	}
	if !ok {
		return errors.ErrNotFound(info.GetObjectTypeName(), path)
	}

	ok, err = vfs.IsDir(fs, path)
	if err != nil {
		return err
	}
	// check directory mode
	if ok {
		data, err := vfs.ReadFile(fs, filepath.Join(path, info.GetDescriptorFileName()))
		if err != nil {
			return err
		}
		return info.ValidateDescriptor(data)
	}

	// check file mode
	ok, err = vfs.IsFile(fs, path)
	if err != nil {
		return err
	}
	if !ok {
		return errors.ErrInvalid(info.GetObjectTypeName(), path)
	}

	r, err := fs.Open(path)
	if err != nil {
		return err
	}
	defer r.Close()

	eff, _, err := compression.AutoDecompress(r)
	if err != nil {
		return err
	}
	defer eff.Close()

	tr := tar.NewReader(eff)
	for {
		header, err := tr.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return fmt.Errorf("no descriptor file (%s) found for %s", info.GetDescriptorFileName(), info.GetObjectTypeName())
			}
			return err
		}

		if header.Typeflag == tar.TypeReg && header.Name == info.GetDescriptorFileName() {
			data, err := io.ReadAll(tr)
			if err != nil {
				return err
			}
			return info.ValidateDescriptor(data)
		}
	}
}
