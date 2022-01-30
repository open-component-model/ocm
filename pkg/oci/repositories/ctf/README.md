
# *Common Transport Format*

The *Common Transport Format* describes a file system structure that can be 
used for the representation of [content](https://github.com/opencontainers/image-spec)
of an OCI repository.

It is a flat directory containing

- **`artefact-index.json`** *[Artefact Index](#artefact-index)*

  This JSON file describes the contained artefact (versions).

- **`artefacts`** *directory*

  The *artefacts* directory contains the [Artefact Archives](#atefact archive-format)
  for the artefact versions described by the artefact descriptor as a flat file
  list. The filename MUST be the digest of the *[Manifest](https://github.com/opencontainers/image-spec/blob/main/manifest.md)*
  of the artefact version. This might be an image or index manifest.

  Hereby the algorithm separator character is replaced by a dot (".").
  Every file SHOULD be referenced in the artefact index. Files not referenced
  here are ignored.

  An archive may be in *tar* or *tgz* format.

It might be used in various technical forms: as structure of an
operating system file system, a virtual file system or as content of
an archive file.

## *Artefact Index*

The *Artefact Index* is a JSON file describing the artefact content in
a file system structucture according to this specification. 

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

  This REQUIRED property is the _digest_ of the targeted artefact, conforming to
  the requirements outlined in
  [Digests](https://github.com/opencontainers/image-spec/blob/main/descriptor.md#digests).
  Retrieved content SHOULD be verified against this digest when consumed via
  untrusted sources.
  
- **`tag`** *string*

  This REQUIRED property is the _tag_ of the targeted artefact, conforming to 
  the requirements outlined in the
  [OCI Distribution Specification](https://github.com/opencontainers/distribution-spec/blob/main/spec.md).
  

## *Artefact Archive* Format

The *Artefact Archive* Format describes a file system structure that can be
used for the representation of a dedicated [artefact version](https://github.com/opencontainers/image-spec)
of an OCI repository. The filesystem content is then packed into a *tar* or *tgz* archive.

It is a directory containing

- **`artefact-descriptor.json`** *[oci artefact manifest](https://github.com/opencontainers/image-spec/blob/main/manifest.md)*

  This JSON file describes the contained artefact (version). It MUST be either
  an image or index manifest.

- **`blobs`** *directory*

  The *blobs* directory contains the blobs described by the artefact descriptor
  as a flat file list. Every file has a filename according to its
  [digest](https://github.com/opencontainers/image-spec/blob/main/descriptor.md#digests). 
  Hereby the algorithm separator character is replaced by a dot (".").
  Every file SHOULD be referenced in the artefact descriptor by a
  [descriptor according the OCI Image Specification](https://github.com/opencontainers/image-spec/blob/main/descriptor.md).

  Files not referenced here are ignored.
