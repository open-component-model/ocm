// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package localociblob

import (
	"fmt"

	. "github.com/open-component-model/ocm/pkg/exception"

	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/runtime"
)

// Type is the access type for a component version local blob in an OCI repository.
const (
	Type   = "localOciBlob"
	TypeV1 = Type + runtime.VersionSeparator + "v1"
)

var versions = cpi.NewAccessTypeVersionScheme(Type)

func init() {
	Must(versions.Register(cpi.NewAccessSpecTypeByConverter(Type, &AccessSpec{}, &converterV1{})))
	Must(versions.Register(cpi.NewAccessSpecTypeByConverter(TypeV1, &AccessSpec{}, &converterV1{})))
	cpi.RegisterAccessTypeVersions(versions)
}

// New creates a new LocalOCIBlob accessor.
// Deprecated: Use LocalBlob.
func New(digest digest.Digest) *localblob.AccessSpec {
	return &localblob.AccessSpec{
		InternalVersionedTypedObject: runtime.NewInternalVersionedTypedObject(versions, Type),
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

func (_ converterV1) ConvertFrom(object cpi.AccessSpec) (runtime.TypedObject, error) {
	in, ok := object.(*localblob.AccessSpec)
	if !ok {
		return nil, fmt.Errorf("failed to assert type %T to localblob.AccessSpec", object)
	}
	return &AccessSpec{
		ObjectVersionedType: runtime.NewVersionedTypedObject(in.Type),
		Digest:              digest.Digest(in.LocalReference),
	}, nil
}

func (_ converterV1) ConvertTo(object interface{}) (cpi.AccessSpec, error) {
	in, ok := object.(*AccessSpec)
	if !ok {
		return nil, fmt.Errorf("failed to assert type %T to localfsblob.AccessSpec", object)
	}
	return &localblob.AccessSpec{
		InternalVersionedTypedObject: runtime.NewInternalVersionedTypedObject(versions, in.Type),
		LocalReference:               in.Digest.String(),
		MediaType:                    "",
	}, nil
}
