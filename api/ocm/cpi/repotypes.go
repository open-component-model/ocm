package cpi

// this file is similar to contexts oci.

import (
	"ocm.software/ocm/api/ocm/compdesc"
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

type repositoryType struct {
	runtime.VersionedTypedObjectType[RepositorySpec]
	checker RepositoryAccessMethodChecker
}

type RepositoryAccessMethodChecker func(Context, compdesc.AccessSpec) bool

func NewRepositoryType[I RepositorySpec](name string, checker RepositoryAccessMethodChecker) RepositoryType {
	return &repositoryType{
		VersionedTypedObjectType: runtime.NewVersionedTypedObjectType[RepositorySpec, I](name),
		checker:                  checker,
	}
}

func NewRepositoryTypeByConverter[I RepositorySpec, V runtime.VersionedTypedObject](name string, converter runtime.Converter[I, V], checker RepositoryAccessMethodChecker) RepositoryType {
	return &repositoryType{
		VersionedTypedObjectType: runtime.NewVersionedTypedObjectTypeByConverter[RepositorySpec, I, V](name, converter),
		checker:                  checker,
	}
}

func NewRepositoryTypeByFormatVersion(name string, fmt runtime.FormatVersion[RepositorySpec], checker RepositoryAccessMethodChecker) RepositoryType {
	return &repositoryType{
		VersionedTypedObjectType: runtime.NewVersionedTypedObjectTypeByFormatVersion[RepositorySpec](name, fmt),
		checker:                  checker,
	}
}

func (t *repositoryType) LocalSupportForAccessSpec(ctx Context, a compdesc.AccessSpec) bool {
	if t.checker != nil {
		return t.checker(ctx, a)
	}
	return false
}
