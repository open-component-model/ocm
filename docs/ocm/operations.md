# Abstract Operations on OCM Repositories

The Open Component Model specification does not describe a dedicated
remotely accessible repository API (like for example the [OCI distribution
specification](https://github.com/opencontainers/distribution-spec/blob/main/spec.md)).

The model is intended to be stored in any kind of storage sub system, which
is able to store a potentially unlimited number of blobs with an adequate
addressing scheme, supporting arbitrary names.

For example, an OCI repository with a deep repository structure, is suitable
to host OCM components (see [OCI mapping Scheme](modelmapping/oci/README.md)).

On the client side a suitable implementation or language binding must available
to work with component information stored in such a storage backend.

The OCM project provides a complete implementation for common OCI registries,
and mapping specification for S3 and OCI.

Every such binding must support at least a dedicated set of abstract operations
working with {elements of the component model}(model.md).

The following operations are mandatory:

- **`UploadComponentDescriptor(ComponentDescriptor-YAML) error`**

  Persist a serialized form of the descriptor of a [component version](model.md#component-versions)  with its
  component identity and version name in way so that is retrievable again using
  this identity.

- **`GetComponentDescriptor(ComponentId, VersionName) (ComponentDescriptor-YAML, error)`**

  Retrieve a formally persisted description of a component version.

- **`UploadBlob(ComponentId, VersionName, BlobAccess, MediaType, ReferenceHint) (BlobIdentity, GlobalAccessSpec, error)`**

  Store a byte stream or blob under a namespace given by the component version
  identity and return a local blob identity (as string) that can be used to retrieve
  the blob, again (together with the component version identity)

  Additionally, a dedicated media type can be used to decide how to internally
  represent the artefact content.

  Optionally the operation may decide to store the blob in dedicated ways according
  to its media type. For example, an OCI based implementation can represent 
  blobs containing an OCI artefact as regular, globally addressable object.

  An type-specific optional *ReferenceHint* can be passed to guide the
  operation for generating an identity, if it decided to make the object
  externally visible.

  If this is the case, an external [access specification](model.md#artefact-access)
  has to be returned. At least a blob identity or an external access specification
  has to be returned for not successful executions.

- **`GetBlob(ComponentId, VersionName, BlobIdentity) (Blob, error)`**

  Retrieve a formerly stored blob, again, using the blob identity provided 
  by the store operation. Technically this should be a stream or the blob content.

- **`ListComponentVersions(ComponentId) ([]VersionName, error)`**

  List all the known versions of a component specified by its component identity.

Optional operations might be:

- **`DeleteComponentVersion(ComponentId, VersionName) error`**

  To be able to clean up old information an operation to delete the information
  stored for a component version should be available.

- **`DeleteBlob(ComponentId, VersionName, BlobIndentity) error`**

  It might be useful to provide an explicit delete operation for blobs stored
  along with the component version. But the repository implementation
  may keep track of used blobs on its own.

- **`ListComponents(ComponentId-Prefix) ([]ComponentId, error)`**

  List all components in the given identifier namespace. (The structure of a 
  component id based on hierarchical namespace)

- **`ListComponents(ComponentId-Prefix) ([]ComponentId, error)`**

  List all components in the given identifier namespace, recursively.
  It should not only return component identities, that are direct children,
  traverse the complete subtree.
