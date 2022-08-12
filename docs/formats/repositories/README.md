# Open Component Model Repository Types

Repository specifications describe (and finally implement) dedicated technical
ways how to access OCM content stored in some technical storage backend.
They are types. A repository type specifies the specification format of a 
repository specification and the technical procedure how to access the 
OCM model content in an instance of this repository determined by the
information store in the specification.

## Specification

Every repository specification has a [type](../../names/repositorytypes.md) and a
specification version. The specification version defines the attribute set
used to describe the information required by the implementation to
address the repository instance.

A repository specification is a yaml/json data fragment with
the following fields:

- **`type`** *string*

  This field describes the repository type and optional specification
  version according to the [repository type naming scheme](../../names/repositorytypes.md)

Additional fields are used to provide the type specific specification.
The fields may have any deep structure.

## Centrally defined Repository Types

The following resource types are centrally defined:

- [`OCIRegistry`](../../../pkg/contexts/ocm/repositories/ocireg/README.md): OCM Repository stored in an OCI registry
- [`CommonTransportFormat`](../../../pkg/contexts/ocm/repositories/ctf/README.md): OCM Repository storing content in the filesystem
- [`ComponentArchive`](../../../pkg/contexts/ocm/repositories/comparch/README.md): Limited OCM Repository hosting a single Component version in the filesystem.

Centrally defined repository types are supported by the 
OCM library and tool set. 