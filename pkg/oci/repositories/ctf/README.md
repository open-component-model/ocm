
# *Common Transport Format*

The *Common Transport Format* describes a file system structure that can be 
used for the representation of [content](https://github.com/opencontainers/image-spec)
of an OCI repository.

It is a directory containing

- **`artefact-index.json`** *[Artefact Index](#artefact-index)*

  This JSON file describes the contained artefact (versions).

- **`blobs`** *directory*

  The *blobs* directory contains the blobs described artefacts described by the
  artefact set descriptor as a flat file list. Every file has a filename according
  to its [digest](https://github.com/opencontainers/image-spec/blob/main/descriptor.md#digests).
  Hereby the algorithm separator character is replaced by a dot (".").
  Every file SHOULD be referenced in the artefact descriptor by a
  [descriptor according the OCI Image Specification](https://github.com/opencontainers/image-spec/blob/main/descriptor.md).

  Files not referenced by the artefacts described by the index are ignored.
  

It might be used in various technical forms: as structure of an
operating system file system, a virtual file system or as content of
an archive file. The descriptor SHOULD be the first file if stored in an archive.

## *Artefact Index*

The *Artefact Index* is a JSON file describing the artefact content in
a file system structure according to this specification. 

### *Artefact Index* Property Descriptions

It contains the following properties.

- **`schemaVersion`** *int*

  This REQUIRED property specifies the index schema version.
  For this version of the specification, this MUST be `1`. The value of this
  field will not change. This field MAY be removed in a future version of the 
  specification.

- **`index`** *[artefact](#artefact-property-descriptions)*


### *Artefact* Property Descriptions

An artefact consists of a set of properties encapsulated in key-value fields.

The following fields contain the properties that constitute an *Artefact*:

- **`repository`** *string*

  This REQUIRED property is the repository name of the targeted artefact described by the
  *Common Transport Format*,  conforming to the requirements outlined in the
  [OCI Distribution Specification](https://github.com/opencontainers/distribution-spec/blob/main/spec.md).

- **`digest`** *string*

  This REQUIRED property is the _digest_ of the targeted artefact in the targeted
  artefact set, conforming to the requirements outlined in
  [Digests](https://github.com/opencontainers/image-spec/blob/main/descriptor.md#digests).
  Retrieved content SHOULD be verified against this digest when consumed via
  untrusted sources.

- **`tag`** *string*

  This optional property is the _tag_ of the targeted artefact, conforming to 
  the requirements outlined in the
  [OCI Distribution Specification](https://github.com/opencontainers/distribution-spec/blob/main/spec.md).

  There might be multiple entries in the artefact list refering to the same artefact
  with different tags. But all used tags for a repository must be unique.
  

## *Artefact Set Archive* Format

The *Artefact Set Archive* Format describes a file system structure that can be
used for the representation of a dedicated set of [artefact versions](https://github.com/opencontainers/image-spec)
of an OCI repository for the same respository.
In the archive form the artefact set descriptor SHOULD be the first file.

The file structure is a directory containing

- **`artefact-set-descriptor.json`** *[oci image index](https://github.com/opencontainers/image-spec/blob/main/image-index.md)*

  This JSON file describes the contained artefact (version). It MUST be an index manifest.
  It MUST describe all *[artefacts](https://github.com/opencontainers/image-spec/blob/main/manifest.md)*
  that should be addressable.

- **`blobs`** *directory*

  The *blobs* directory contains the blobs described by the artefact set descriptor
  as a flat file list. Every file has a filename according to its
  [digest](https://github.com/opencontainers/image-spec/blob/main/descriptor.md#digests). 
  Hereby the algorithm separator character is replaced by a dot (".").
  Every file SHOULD be referenced in the artefact descriptor by a
  [descriptor according the OCI Image Specification](https://github.com/opencontainers/image-spec/blob/main/descriptor.md).

  Files not referenced by the artefacts described by the index are ignored.

The artefact set index describes the manifest entries that should be registered
via the manifest endpoint according to the distribution spec.

### Extension Models

Any artefacts described by this format is addressable only by digests.
The entries in the index are intended to be registered via the manifest
endpoint according to the [OCI Distribution Specification](https://github.com/opencontainers/distribution-spec).

Additionally this specification described two special annotations:

- **`ocm.gardener.cloud/tags`**
  
  This annotation can be used to describe a comma-separated list of tags.
  that should be added for this index entry, when imported into an OCI registry.

- **`ocm.gardener.cloud/type`**

  This annotation can be used some additional type information.

This way the format can be used to attach elements according to various extension
models for the OCI specification:

 - *[cosign](https://github.com/sigstore/cosign)*

   For *cosign* addtional signatures are represented as dedicated artefacts
   with special tags establishing the relation to the original artefact they
   refer to by deriving the tag from the digest of the related object.

   This is supported by this format by providing a possibility to describe
   additional tags by an annotation in the index

 - [*ORAS*](https://github.com/oras-project/artifacts-spec)

   Here a new third top-level manifest type is introduced, that can be 
   stored via the manifest endpoint of the distribution spec. No addtional
   tags are required, the relation to the annotated object is established
   by a dedicated digest based field. Those artefacts can directly be 
   described by this format. But language bindings basically have to support
   this additional type.