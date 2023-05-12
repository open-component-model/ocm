// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cpi

// this file is identical for contexts oci and credentials and similar for
// ocm.

import (
	"github.com/open-component-model/ocm/pkg/runtime"
)

type RepositoryTypeVersionScheme = runtime.TypeVersionScheme[RepositorySpec, RepositoryType]

func NewRepositoryTypeVersionScheme(kind string) RepositoryTypeVersionScheme {
	return runtime.NewTypeVersionScheme[RepositorySpec, RepositoryType](kind, newStrictRepositoryTypeScheme())
}

func RegisterRepositoryType(rtype RepositoryType) {
	defaultRepositoryTypeScheme.Register(rtype)
}

func RegisterRepositoryTypeVersions(s RepositoryTypeVersionScheme) {
	defaultRepositoryTypeScheme.AddKnownTypes(s)
}

////////////////////////////////////////////////////////////////////////////////

func NewRepositoryType[I RepositorySpec](name string, proto I) RepositoryType {
	return runtime.NewVersionedTypedObjectTypeByProto[RepositorySpec, I](name, proto)
}

func NewRepositoryTypeByProtoConverter[I RepositorySpec](name string, proto runtime.TypedObject, converter runtime.Converter[I, runtime.TypedObject]) RepositoryType {
	return runtime.NewVersionedTypedObjectTypeByProtoConverter[RepositorySpec, I](name, proto, converter)
}

func NewRepositoryTypeByConverter[I RepositorySpec, V runtime.TypedObject](name string, converter runtime.Converter[I, V]) RepositoryType {
	return runtime.NewVersionedTypedObjectTypeByConverter[RepositorySpec, I](name, converter)
}
