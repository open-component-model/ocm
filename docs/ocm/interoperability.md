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
implemention of the mapping schemes required for the storage technologies
intended to be used in this environment.

## Support Library

The standard Go implementation provided by the project, provides a general
library supporting mappings for
- a filesystem representation
- an OCI based representation

Based on this library, a command line client tool is provided that uses
those bindings in a generic way to support typical tasks to work with
the component model.

This kind of generalized interface for the library consumer is extensible and
makes it possible
to add further mapping implementations just be registering new
type objects with function provided by the library.

## Mapping Specifications

Mappings are lready specifid for the following storage backends:

- [**`CommonTransportFormat`**](modelmapping/filesystem/README.md) File structure stored under
  a directory or in tar or tgz file.

- [**`OCIRegistry`**](modelmapping/oci/README.md) storing OCM component versions in
  an OCI conform registry

- [**`S3`**](modelmapping/s3/README.md) storing OCM component versions in an S3 blob
  store