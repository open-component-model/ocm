# Open Component Model Repository Types

Repository specifications describe technical
ways of how to access OCM content stored in a technical storage backend of
a dedicated type. To distinguish specifications for different types of
storage backends they are typed. A repository type defines

- the specification format of a repository specification used to describe a 
  dedicated repository instance.
- the technical procedure how to access the OCM model content in an instance of
  this repository determined by the information stored in such a repository
  specification.

## Specification

Every repository specification has a [type](../../names/repositorytypes.md) and
an optional specification version. The specification version defines the attribute
set used to describe the information required by the implementation to
identify and gain access the repository instance (without access credentials).
The default specification version is `v1`.

A repository specification is a yaml/json data fragment with
the following fields:

- **`type`** *string*

  This field describes the repository type and optional specification
  version according to the [repository type naming scheme](../../names/repositorytypes.md).

Every repository type can define arbitrary additional fields.
These fields may have any deep, but type specific and defined structure.
The type defines this attribute structure and its interpretation is left to the
concrete implementation of the repository type. All implementations of a
dedicated type have to conform to the attribute structure definition of this type.

For example, the type `OCIRegistry` defines two additional flat
[attributes](../../../pkg/contexts/ocm/repositories/ocireg/README.md): `baseURL`
and `legacyTypes`.

## Centrally defined Repository Types

The following repository types are centrally defined:

- [`OCIRegistry`](../../../pkg/contexts/ocm/repositories/ocireg/README.md): OCM Repository storing content in an OCI registry
- [`CommonTransportFormat`](../../../pkg/contexts/ocm/repositories/ctf/README.md): OCM Repository storing content on the filesystem
- [`ComponentArchive`](../../../pkg/contexts/ocm/repositories/comparch/README.md): Limited OCM Repository capable to store a single [component version](../../ocm/model.md#component-versions) on the filesystem

Centrally defined repository types are supported by the 
OCM library and tool set. 