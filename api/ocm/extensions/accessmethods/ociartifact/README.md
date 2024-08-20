
# Access Method `ociArtifact` and `ociRegistry` - OCI Artifact Access

## Synopsis

```yaml
type: ociArtifact/v1
```

Provided blobs use the following media type:

- `application/vnd.oci.image.manifest.v1+tar+gzip`: OCI image manifests
- `application/vnd.oci.image.index.v1+tar.gzip`: OCI index manifests

Depending on the repository appropriate docker legacy types might be used.

The artifact content is provided in the [Artifact Set Format](/api/oci/extensions/repositories/ctf/formatspec.md#artifact-set-archive-format).
The tag is provided as annotation.

### Description

This method implements the access of an OCI artifact stored in an OCI registry.

Supported specification version is `v1`

### Specification Versions

#### Version `v1`

The type specific specification fields are:

- **`imageReference`** *string*

  OCI image/artifact reference following the possible docker schemes:
  - `<repo>/<artifact>:<tag>@<digest>`
  - `<host>[<port>]/<repo path>/<artifact>:<tag>@<digest>`

### Go Bindings

The go binding can be found [here](method.go)
