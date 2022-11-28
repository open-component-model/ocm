# Mapping Specifications for OCM Storage Backends

The operational specification of the Open Component Model restricts itself
to an abstract set of operations an [OCM repository](model.md#repositories)
must support.

It does not describe a dedicated
remotely accessible repository API (like for example the [OCI distribution
specification](https://github.com/opencontainers/distribution-spec/blob/main/spec.md)).

The model is intended to be stored in any kind of storage sub system, which
is able to store a potentially unlimited number of blobs with an adequate
addressing scheme, supporting arbitrary names.

This is achieved by specifying mapping schemes for every supported storage 
backend technology, which describes how elements of the [component model](model.md)
are mapped to elements provided by the underlying storage technology.

The OCM based tools used in a dedicated environment must use an adequate
implementation of the mapping schemes required for the storage technologies
intended to be used in this environment. It finally has to implement the
[abstract model operations based on this mapping scheme](operations.md)

## Support Library

The standard Go implementation provided by the project, provides a general
library supporting mappings for
- a filesystem representation
- an OCI based representation
  
The library encapsulates those bindings under a generic and extensible interface.
Based on this library, a command line client tool is provided, using this
generalized interface to implement typical tasks to work with the component model
in a repository type agnostic way.

This kind of generalized interface for the library consumer is extensible and
makes it possible
to add further mapping implementations just by registering new
type objects with functions provided by the library. So, combining this
library, the new/local type implementations and the CLI in a new executable,
can extend the existing generic CLI functionality with the supported new
repository types.

There are several such extension points:
- repository types
- access method types
- specialized artifact down-loaders (for example foe helm charts)
- specialized artifact digesters (for example for OCI images)
- transport policies
- credential repositories and types
- signing handlers

The provided library already supports a usable and consistent set
of those elements to work with filesystem and OCI registries with 
dedicated specialized support for helm charts and OCI images. 

## Mapping Specifications

Mappings are already specified for the following storage backends:

- [**`CommonTransportFormat`**](modelmapping/filesystem/README.md) File structure stored under
  a directory or in tar or tgz file.

- [**`OCIRegistry`**](modelmapping/oci/README.md) storing OCM component versions in
  an OCI conform registry

- [**`S3`**](modelmapping/s3/README.md) storing OCM component versions in an S3 blob
  store