package localfsblob

import (
	. "github.com/mandelsoft/goutils/exception"
	"ocm.software/ocm/api/ocm/cpi/accspeccpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/utils/runtime"
)

// Type is the access type of a blob in a local filesystem.
const (
	Type   = "localFilesystemBlob"
	TypeV1 = Type + runtime.VersionSeparator + "v1"
)

// Keep old access method and map generic one to this implementation for component archives

// This method uses the localblob internal format and converts it to/from the
// appropriate serialization version.
// The attributes referenceName and globalAccess are NOT supported.

var versions = accspeccpi.NewAccessTypeVersionScheme(Type)

func init() {
	Must(versions.Register(accspeccpi.NewAccessSpecTypeByConverter[*localblob.AccessSpec, *AccessSpec](Type, &converterV1{})))
	Must(versions.Register(accspeccpi.NewAccessSpecTypeByConverter[*localblob.AccessSpec, *AccessSpec](TypeV1, &converterV1{})))
	accspeccpi.RegisterAccessTypeVersions(versions)
}

// New creates a new localFilesystemBlob accessor.
func New(path string, media string) *localblob.AccessSpec {
	return &localblob.AccessSpec{
		InternalVersionedTypedObject: runtime.NewInternalVersionedTypedObject[accspeccpi.AccessSpec](versions, Type),
		LocalReference:               path,
		MediaType:                    media,
	}
}

func Decode(data []byte) (*localblob.AccessSpec, error) {
	spec, err := versions.Decode(data, runtime.DefaultYAMLEncoding)
	if err != nil {
		return nil, err
	}
	return spec.(*localblob.AccessSpec), nil
}

// AccessSpec describes the access for a blob on the filesystem.
// Deprecated: use LocalBlob.
type AccessSpec struct {
	runtime.ObjectVersionedType `json:",inline"`
	// FileName is the
	Filename string `json:"fileName"`
	// MediaType is the media type of the object represented by the blob
	MediaType string `json:"mediaType"`
}

////////////////////////////////////////////////////////////////////////////////

type converterV1 struct{}

func (_ converterV1) ConvertFrom(in *localblob.AccessSpec) (*AccessSpec, error) {
	return &AccessSpec{
		ObjectVersionedType: runtime.NewVersionedTypedObject(in.Type),
		Filename:            in.LocalReference,
		MediaType:           in.MediaType,
	}, nil
}

func (_ converterV1) ConvertTo(in *AccessSpec) (*localblob.AccessSpec, error) {
	return &localblob.AccessSpec{
		InternalVersionedTypedObject: runtime.NewInternalVersionedTypedObject[accspeccpi.AccessSpec](versions, in.Type),
		LocalReference:               in.Filename,
		MediaType:                    in.MediaType,
	}, nil
}
