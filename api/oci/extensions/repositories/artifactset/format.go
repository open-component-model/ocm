package artifactset

import (
	"path/filepath"
	"strings"
	"sync"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/mime"
)

const (
	// ArtifactSetDescriptorFileName is the artifact descriptor name for artifact format.
	ArtifactSetDescriptorFileName = "artifact-descriptor.json"
	BlobsDirectoryName            = "blobs"

	OCIArtifactSetDescriptorFileName = "index.json"
	OCILayouFileName                 = "oci-layout"
)

var DefaultArtifactSetDescriptorFileName = OCIArtifactSetDescriptorFileName

func IsOCIDefaultFormat() bool {
	return DefaultArtifactSetDescriptorFileName == OCIArtifactSetDescriptorFileName
}

func DescriptorFileName(format string) string {
	switch format {
	case FORMAT_OCI, FORMAT_OCI_COMPLIANT:
		return OCIArtifactSetDescriptorFileName
	case FORMAT_OCM:
		return ArtifactSetDescriptorFileName
	case "":
		return DefaultArtifactSetDescriptorFileName
	}
	return ""
}

type accessObjectInfo struct {
	accessobj.DefaultAccessObjectInfo
	format string // FORMAT_OCI, FORMAT_OCI_COMPLIANT, or FORMAT_OCM
}

var _ accessobj.AccessObjectInfo = (*accessObjectInfo)(nil)

var baseInfo = accessobj.DefaultAccessObjectInfo{
	ObjectTypeName:           "artifactset",
	ElementDirectoryName:     BlobsDirectoryName,
	ElementTypeName:          "blob",
	DescriptorHandlerFactory: NewStateHandler,
	DescriptorValidator:      validateDescriptor,
}

func validateDescriptor(data []byte) error {
	_, err := artdesc.DecodeIndex(data)
	return err
}

func NewAccessObjectInfo(fmts ...string) accessobj.AccessObjectInfo {
	a := &accessObjectInfo{
		DefaultAccessObjectInfo: baseInfo,
	}
	oci := IsOCIDefaultFormat()
	if len(fmts) > 0 {
		switch fmts[0] {
		case FORMAT_OCM:
			oci = false
		case FORMAT_OCI:
			oci = true
		case FORMAT_OCI_COMPLIANT:
			a.format = FORMAT_OCI_COMPLIANT
			oci = true
		case "":
		}
	}
	if oci {
		a.setOCI()
	} else {
		a.setOCM()
	}
	return a
}

func (a *accessObjectInfo) setOCI() {
	a.DescriptorFileName = OCIArtifactSetDescriptorFileName
	a.AdditionalFiles = []string{OCILayouFileName}
}

func (a *accessObjectInfo) setOCM() {
	a.DescriptorFileName = ArtifactSetDescriptorFileName
	a.AdditionalFiles = nil
}

// SubPath returns the path for a blob. For OCI-compliant format, converts
// "sha256.DIGEST" to "blobs/sha256/DIGEST" per OCI Image Layout Specification.
func (a *accessObjectInfo) SubPath(name string) string {
	if a.format == FORMAT_OCI_COMPLIANT {
		// Convert sha256.DIGEST to sha256/DIGEST for OCI compliance
		if algo, dig, ok := strings.Cut(name, "."); ok {
			return filepath.Join(a.ElementDirectoryName, algo, dig)
		}
	}
	return filepath.Join(a.ElementDirectoryName, name)
}

func (a *accessObjectInfo) setupOCIFS(fs vfs.FileSystem, mode vfs.FileMode) error {
	data := `{
    "imageLayoutVersion": "1.0.0"
}
`
	return vfs.WriteFile(fs, OCILayouFileName, []byte(data), mode)
}

func (a *accessObjectInfo) SetupFileSystem(fs vfs.FileSystem, mode vfs.FileMode) error {
	if err := a.SetupFor(fs); err != nil {
		return err
	}
	if err := a.DefaultAccessObjectInfo.SetupFileSystem(fs, mode); err != nil {
		return err
	}
	if len(a.AdditionalFiles) > 0 {
		return a.setupOCIFS(fs, mode)
	}
	return nil
}

func (a *accessObjectInfo) SetupFor(fs vfs.FileSystem) error {
	ok, err := vfs.FileExists(fs, OCIArtifactSetDescriptorFileName)
	if err != nil {
		return err
	}
	if ok {
		a.setOCIDescriptor()
		return nil
	}

	ok, err = vfs.FileExists(fs, ArtifactSetDescriptorFileName)
	if err != nil {
		return err
	}
	if ok {
		a.setOCM()
		return nil
	}

	ok, err = vfs.FileExists(fs, OCILayouFileName)
	if err != nil {
		return err
	}
	if ok {
		a.setOCIDescriptor()
		return nil
	}

	// keep configured format
	return nil
}

// setOCIDescriptor sets OCI descriptor/files but preserves format
// (which determines flat vs nested blob paths based on FORMAT_OCI vs FORMAT_OCI_COMPLIANT).
func (a *accessObjectInfo) setOCIDescriptor() {
	a.DescriptorFileName = OCIArtifactSetDescriptorFileName
	a.AdditionalFiles = []string{OCILayouFileName}
}

////////////////////////////////////////////////////////////////////////////////

type Object = ArtifactSet

type FormatHandler interface {
	accessio.Option

	Format() accessio.FileFormat

	Open(acc accessobj.AccessMode, path string, opts accessio.Options) (*Object, error)
	Create(path string, opts accessio.Options, mode vfs.FileMode) (*Object, error)
	Write(obj *Object, path string, opts accessio.Options, mode vfs.FileMode) error
}

type formatHandler struct {
	accessobj.FormatHandler
}

var (
	FormatDirectory = RegisterFormat(accessobj.FormatDirectory)
	FormatTAR       = RegisterFormat(accessobj.FormatTAR)
	FormatTGZ       = RegisterFormat(accessobj.FormatTGZ)
)

////////////////////////////////////////////////////////////////////////////////

var (
	fileFormats = map[accessio.FileFormat]FormatHandler{}
	lock        sync.RWMutex
)

func RegisterFormat(f accessobj.FormatHandler) FormatHandler {
	lock.Lock()
	defer lock.Unlock()
	h := &formatHandler{f}
	fileFormats[f.Format()] = h
	return h
}

func GetFormats() []string {
	lock.RLock()
	defer lock.RUnlock()
	return accessio.GetFormatsFor(fileFormats)
}

func GetFormat(name accessio.FileFormat) FormatHandler {
	lock.RLock()
	defer lock.RUnlock()
	return fileFormats[name]
}

func SupportedFormats() []accessio.FileFormat {
	lock.RLock()
	defer lock.RUnlock()
	result := make([]accessio.FileFormat, 0, len(fileFormats))
	for f := range fileFormats {
		result = append(result, f)
	}
	return result
}

////////////////////////////////////////////////////////////////////////////////

func OpenFromBlob(acc accessobj.AccessMode, blob blobaccess.BlobAccess, opts ...accessio.Option) (*Object, error) {
	return OpenFromDataAccess(acc, blob.MimeType(), blob, opts...)
}

func OpenFromDataAccess(acc accessobj.AccessMode, mediatype string, data blobaccess.DataAccess, opts ...accessio.Option) (*Object, error) {
	o, err := accessio.AccessOptions(nil, opts...)
	if err != nil {
		return nil, err
	}
	if o.GetFile() != nil || o.GetReader() != nil {
		return nil, errors.ErrInvalid("file or reader option not possible for blob access")
	}
	reader, err := data.Reader()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	o.SetReader(reader)
	fmt := accessio.FormatTar

	if mime.IsGZip(mediatype) {
		fmt = accessio.FormatTGZ
	}
	o.SetFileFormat(fmt)
	return Open(acc&accessobj.ACC_READONLY, "", 0, o)
}

func Open(acc accessobj.AccessMode, path string, mode vfs.FileMode, olist ...accessio.Option) (*Object, error) {
	o, create, err := accessobj.HandleAccessMode(acc, path, &Options{}, olist...)
	if err != nil {
		return nil, err
	}
	h, ok := fileFormats[*o.GetFileFormat()]
	if !ok {
		return nil, errors.ErrUnknown(accessobj.KIND_FILEFORMAT, o.GetFileFormat().String())
	}
	if create {
		return h.Create(path, o, mode)
	}
	return h.Open(acc, path, o)
}

func Create(acc accessobj.AccessMode, path string, mode vfs.FileMode, opts ...accessio.Option) (*Object, error) {
	o, err := accessio.AccessOptions(&Options{}, opts...)
	if err != nil {
		return nil, err
	}
	o.DefaultFormat(accessio.FormatDirectory)
	h, ok := fileFormats[*o.GetFileFormat()]
	if !ok {
		return nil, errors.ErrUnknown(accessobj.KIND_FILEFORMAT, o.GetFileFormat().String())
	}
	return h.Create(path, o, mode)
}

////////////////////////////////////////////////////////////////////////////////

func (h *formatHandler) Open(acc accessobj.AccessMode, path string, opts accessio.Options) (*Object, error) {
	return _Wrap(h.FormatHandler.Open(NewAccessObjectInfo(GetFormatVersion(opts)), acc, path, opts))
}

func (h *formatHandler) Create(path string, opts accessio.Options, mode vfs.FileMode) (*Object, error) {
	return _Wrap(h.FormatHandler.Create(NewAccessObjectInfo(GetFormatVersion(opts)), path, opts, mode))
}

// WriteToFilesystem writes the current object to a filesystem.
func (h *formatHandler) Write(obj *Object, path string, opts accessio.Options, mode vfs.FileMode) error {
	return h.FormatHandler.Write(obj.container.base.Access(), path, opts, mode)
}
