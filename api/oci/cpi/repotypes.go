package cpi

// this file is identical for contexts oci and credentials and similar for
// ocm.

import (
	"ocm.software/ocm/api/utils/runtime"
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

func NewRepositoryType[I RepositorySpec](name string) RepositoryType {
	return runtime.NewVersionedTypedObjectType[RepositorySpec, I](name)
}

func NewRepositoryTypeByConverter[I RepositorySpec, V runtime.TypedObject](name string, converter runtime.Converter[I, V]) RepositoryType {
	return runtime.NewVersionedTypedObjectTypeByConverter[RepositorySpec, I](name, converter)
}

func NewRepositoryTypeByFormatVersion(name string, fmt runtime.FormatVersion[RepositorySpec]) RepositoryType {
	return runtime.NewVersionedTypedObjectTypeByFormatVersion[RepositorySpec](name, fmt)
}
