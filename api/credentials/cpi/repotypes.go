package cpi

// this file is identical for contexts oci and credentials and similar for
// ocm.

import (
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/api/utils/runtime/descriptivetype"
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

func NewRepositoryType[I RepositorySpec](name string, opts ...RepositoryOption) RepositoryType {
	return descriptivetype.NewVersionedTypedObjectTypeObject(runtime.NewVersionedTypedObjectType[RepositorySpec, I](name), opts...)
}

func NewRepositoryTypeByConverter[I RepositorySpec, V runtime.TypedObject](name string, converter runtime.Converter[I, V], opts ...RepositoryOption) RepositoryType {
	return descriptivetype.NewVersionedTypedObjectTypeObject(runtime.NewVersionedTypedObjectTypeByConverter[RepositorySpec, I](name, converter), opts...)
}

func NewRepositoryTypeByFormatVersion(name string, fmt runtime.FormatVersion[RepositorySpec], opts ...RepositoryOption) RepositoryType {
	return descriptivetype.NewVersionedTypedObjectTypeObject(runtime.NewVersionedTypedObjectTypeByFormatVersion[RepositorySpec](name, fmt), opts...)
}

////////////////////////////////////////////////////////////////////////////////

type RepositoryOption = descriptivetype.Option

func WithDescription(v string) RepositoryOption {
	return descriptivetype.WithDescription(v)
}

func WithFormatSpec(v string) RepositoryOption {
	return descriptivetype.WithFormatSpec(v)
}
