package localociblob

import (
	. "github.com/mandelsoft/goutils/exception"

	"github.com/opencontainers/go-digest"

	"ocm.software/ocm/api/ocm/cpi/accspeccpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/utils/runtime"
)

// Type is the access type for a component version local blob in an OCI repository.
const (
	Type   = "localOciBlob"
	TypeV1 = Type + runtime.VersionSeparator + "v1"
)

var versions = accspeccpi.NewAccessTypeVersionScheme(Type)

func init() {
	Must(versions.Register(accspeccpi.NewAccessSpecTypeByConverter[*localblob.AccessSpec, *AccessSpec](Type, &converterV1{})))
	Must(versions.Register(accspeccpi.NewAccessSpecTypeByConverter[*localblob.AccessSpec, *AccessSpec](TypeV1, &converterV1{})))
	accspeccpi.RegisterAccessTypeVersions(versions)
}

// New creates a new LocalOCIBlob accessor.
// Deprecated: Use LocalBlob.
func New(digest digest.Digest) *localblob.AccessSpec {
	return &localblob.AccessSpec{
		InternalVersionedTypedObject: runtime.NewInternalVersionedTypedObject[accspeccpi.AccessSpec](versions, Type),
		LocalReference:               digest.String(),
	}
}

func Decode(data []byte) (*localblob.AccessSpec, error) {
	spec, err := versions.Decode(data, runtime.DefaultYAMLEncoding)
	if err != nil {
		return nil, err
	}
	return spec.(*localblob.AccessSpec), nil
}

// AccessSpec describes the access for a oci registry.
// Deprecated: Use LocalBlob.
type AccessSpec struct {
	runtime.ObjectVersionedType `json:",inline"`

	// Digest is the digest of the targeted content.
	Digest digest.Digest `json:"digest"`
}

////////////////////////////////////////////////////////////////////////////////

type converterV1 struct{}

func (_ converterV1) ConvertFrom(in *localblob.AccessSpec) (*AccessSpec, error) {
	return &AccessSpec{
		ObjectVersionedType: runtime.NewVersionedTypedObject(in.Type),
		Digest:              digest.Digest(in.LocalReference),
	}, nil
}

func (_ converterV1) ConvertTo(in *AccessSpec) (*localblob.AccessSpec, error) {
	return &localblob.AccessSpec{
		InternalVersionedTypedObject: runtime.NewInternalVersionedTypedObject[accspeccpi.AccessSpec](versions, in.Type),
		LocalReference:               in.Digest.String(),
		MediaType:                    "",
	}, nil
}
