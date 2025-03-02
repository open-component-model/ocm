package comparch

import (
	"sync"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
)

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
// ComponentDescriptorFileName is the name of the component-descriptor file.
const ComponentDescriptorFileName = compdesc.ComponentDescriptorFileName

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
// BlobsDirectoryName is the name of the blob directory in the tar.
const BlobsDirectoryName = "blobs"

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
var accessObjectInfo = &accessobj.DefaultAccessObjectInfo{
	DescriptorFileName:       ComponentDescriptorFileName,
	ObjectTypeName:           "component archive",
	ElementDirectoryName:     BlobsDirectoryName,
	ElementTypeName:          "blob",
	DescriptorHandlerFactory: NewStateHandler,
	DescriptorValidator:      validateDescriptor,
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func validateDescriptor(data []byte) error {
	_, err := compdesc.Decode(data)
	return err
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
type Object = ComponentArchive

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
type FormatHandler interface {
	accessio.Option

	Format() accessio.FileFormat

	Open(ctx cpi.ContextProvider, acc accessobj.AccessMode, path string, opts accessio.Options) (*Object, error)
	Create(ctx cpi.ContextProvider, path string, opts accessio.Options, mode vfs.FileMode) (*Object, error)
	Write(obj *Object, path string, opts accessio.Options, mode vfs.FileMode) error
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
type formatHandler struct {
	accessobj.FormatHandler
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
var (
	FormatDirectory = RegisterFormat(accessobj.FormatDirectory)
	FormatTAR       = RegisterFormat(accessobj.FormatTAR)
	FormatTGZ       = RegisterFormat(accessobj.FormatTGZ)
)

////////////////////////////////////////////////////////////////////////////////

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
var (
	fileFormats = map[accessio.FileFormat]*formatHandler{}
	lock        sync.RWMutex
)

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func RegisterFormat(f accessobj.FormatHandler) *formatHandler {
	lock.Lock()
	defer lock.Unlock()
	h := &formatHandler{f}
	fileFormats[f.Format()] = h
	return h
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func GetFormats() []string {
	lock.RLock()
	defer lock.RUnlock()
	return accessio.GetFormatsFor(fileFormats)
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func GetFormat(name accessio.FileFormat) FormatHandler {
	lock.RLock()
	defer lock.RUnlock()
	h, ok := fileFormats[name]
	if ok {
		return h
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func Open(ctx cpi.ContextProvider, acc accessobj.AccessMode, path string, mode vfs.FileMode, olist ...accessio.Option) (*Object, error) {
	opts, err := accessio.AccessOptions(&accessio.StandardOptions{PathFileSystem: vfsattr.Get(ctx.OCMContext())}, olist...)
	if err != nil {
		return nil, err
	}
	o, create, err := accessobj.HandleAccessMode(acc, path, opts)
	if err != nil {
		return nil, err
	}
	h, ok := fileFormats[*o.GetFileFormat()]
	if !ok {
		return nil, errors.ErrUnknown(accessobj.KIND_FILEFORMAT, o.GetFileFormat().String())
	}
	if create {
		return h.Create(ctx, path, o, mode)
	}
	return h.Open(ctx, acc, path, o)
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func Create(ctx cpi.ContextProvider, acc accessobj.AccessMode, path string, mode vfs.FileMode, opts ...accessio.Option) (*Object, error) {
	o, err := accessio.AccessOptions(nil, opts...)
	if err != nil {
		return nil, err
	}
	o.DefaultFormat(accessio.FormatDirectory)
	h, ok := fileFormats[*o.GetFileFormat()]
	if !ok {
		return nil, errors.ErrUnknown(accessobj.KIND_FILEFORMAT, o.GetFileFormat().String())
	}
	return h.Create(ctx, path, o, mode)
}

////////////////////////////////////////////////////////////////////////////////

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (h *formatHandler) Open(ctx cpi.ContextProvider, acc accessobj.AccessMode, path string, opts accessio.Options) (*Object, error) {
	obj, err := h.FormatHandler.Open(accessObjectInfo, acc, path, opts)
	if err != nil {
		return nil, err
	}
	spec, err := NewRepositorySpec(acc, path, opts)
	return _Wrap(ctx, obj, spec, err)
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (h *formatHandler) Create(ctx cpi.ContextProvider, path string, opts accessio.Options, mode vfs.FileMode) (*Object, error) {
	obj, err := h.FormatHandler.Create(accessObjectInfo, path, opts, mode)
	if err != nil {
		return nil, err
	}
	spec, err := NewRepositorySpec(accessobj.ACC_CREATE, path, opts)
	return _Wrap(ctx, obj, spec, err)
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
// WriteToFilesystem writes the current object to a filesystem.
func (h *formatHandler) Write(obj *Object, path string, opts accessio.Options, mode vfs.FileMode) error {
	return h.FormatHandler.Write(obj.container.fsacc.Access(), path, opts, mode)
}
