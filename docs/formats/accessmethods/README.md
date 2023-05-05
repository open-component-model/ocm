# Access Methods of the Open Component Model

Access methods describe (and finally implement) dedicated technical ways how to
access the blob content of a (re)source described by an
[OCM component descriptor](../../formats/compdesc/README.md).

They are always evaluated in the storage context used to read the component
descriptor containing the access method description. There are two technical
flavors of access methods:

- *external access methods*

  Those methods are self-describing and refer to
  resources stored externally in any supported repository type. 

- *local access methods*

  The access method is evaluated in relation to the repository used to read
  the component descriptor, which described it. So, this content is potentially
  read from different repositories, if the component version is moved among
  repositories, although the properties may be unchanged.

Access Methods are an integral part of the Open Component Model. They always 
provide the link between a component version stored in some repository context,
and the technical access of the described resources applicable for this
context. Therefore, the access method of a resource may change when 
component versions are transported among repository contexts.

Because of this feature, their implementation must be known to provide basic
functionality of the model, like transporting
content by value, in all kinds of environment.

In a dedicated context all used access methods must be known by the used tool
set. Nevertheless, the set of access methods is not fixed. The actual
library/tool version provides a simple way to locally add new methods with
their implementations to support own local environments. This is just done
by providing an own main function with anonymous imports to the new
access method packages (example links can be found below each documented
access method). An external plugin mechanism by calling a separate
executable is not yet supported.

## Specification

Every access method has a [type](../../names/accessmethods.md) and a
specification version. The specification version defines the attribute set
used to describe the information required by the implementation to
address the resource blob.

A complete access method specification is a yaml data fragment with
the following fields:

- **`type`** *string*

  This field describes the access method type and optional specification
  version according to the [access method naming scheme](../../names/accessmethods.md)

Additional fields are used to provide the type specific specification.
The fields may have any deep structure.

## Centrally defined Access Methods

The following method types are centrally defined and available in the OCM toolset:

- [`ociArtifact`](../../../pkg/contexts/ocm/accessmethods/ociartifact/README.md) *external*
  
  Access of an OCI artifact stored in any OCI registry.

- [`ociBlob`](../../../pkg/contexts/ocm/accessmethods/ociblob/README.md) *external*

  Access of an OCI artifact stored in any OCI registry.

- [`gitHub`](../../../pkg/contexts/ocm/accessmethods/github/README.md) *external*

  Access of a git commit in a [github](https://github.com) repository.

- [`helm`](../../../pkg/contexts/ocm/accessmethods/helm/README.md) *external*

  Access of a Helm chart in a [helm](https://helm.sh/docs/topics/chart_repository/) repository.

- [`s3`](../../../pkg/contexts/ocm/accessmethods/s3/README.md) *external*

  Access of a blob in an S3 blob store.

- [`npm`](../../../pkg/contexts/ocm/accessmethods/npm/README.md) *external*

  Access of an NPM module from an NPM registry.

- [`localBlob`](../../../pkg/contexts/ocm/accessmethods/localblob/README.md) *local*

  This is a special access method that has no global implementation.
  It is used to store resource blobs along with the component descriptor
  as part of the OCM repository. Therefore, all OCM repository implementations
  MUST explicitly provide an implementation for this method.

  For example, in an OCI based OCM repository the implementation stores
  local blobs as additional artifact layers according to the OCI model.

### Legacy Types

- [`ociRegistry`](../../../pkg/contexts/ocm/accessmethods/ociartifact/README.md) *external* (deprecated)

  It is a legacy name on the new official access method `ociArtifact`

- [`github`](../../../pkg/contexts/ocm/accessmethods/github/README.md) *external*

  It is a legacy name on the new official access method `gitHub`
