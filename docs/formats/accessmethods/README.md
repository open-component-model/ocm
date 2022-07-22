# Access Methods of the Open Component Model

Access methods describe (and finally implement) dedicated technical ways how to
access the blob  content of a (re)source described by an
[OCM component descriptor](../../formats/compdesc/README.md) in the storage
context of the component descriptor containing the access method description.

They are an integral part of the Open Component Model, because their implementation
must be known to provide basic functionality of the model, like transporting
content by value.

In a dedicated context all used access methods must be known by the used tool
set. Nevertheless, the set of access methods is not fixed. The actual
library/tool version provides a simple way to locally add new methods with
their implementations to support own local environments.

## Specification

Every access method has [type](../../names/accessmethods.md) and a
specification version. The specification version defined the attribute set
used to describe the information required by the implementation to
address the resource blob.

A complete access method specification is a yaml data fragment with
the following fields:

- **`type`** *string*

  This field describes the access method type and optional specification
  version according to the [access method naming scheme](../../names/accessmethods.md)

Additional field are used to provide the type specific specification.
The fields may have any deep structure.

## Centrally defined Access Methods

The following method types are centrally defined and available with
this version of the OCM toolset:

- [`ociArtefact`](../../../pkg/contexts/ocm/accessmethods/ociartefact/README.md)
  
  Access of an OCI artefact stored in any OCI registry.

- [`ociRegistry`](../../../pkg/contexts/ocm/accessmethods/ociartefact/README.md) (deprecated)

  It is a legacy name on the new official access method `ociArtefact`

- [`ociBlob`](../../../pkg/contexts/ocm/accessmethods/ociblob/README.md)

  Access of an OCI artefact stored in any OCI registry.

- [`localBlob`](../../../pkg/contexts/ocm/accessmethods/localblob/README.md)

  This is a special access method that has no global implementation.
  It is used to store resource blobs along with the component descriptor
  as part of the OCM repository. Therefore, all OCM repository implementations
  MUST explicitly provide an implementation for this method.

  For example, in an OCI based OCM repository the implementation stores
  local blobs as additional artefact layers according to the OCI model.