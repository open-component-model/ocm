
# Repository `OCIRegistry` and `ociRegistry` - OCI Registry 


### Synopsis

```
type: OCIRegistry/v1
```

### Description

The content of an OCI-like repository will be stored in an OCI repository
according to the [OCI distribution specification](https://github.com/opencontainers/distribution-spec/blob/main/spec.md).

Supported specification version is `v1`



### Specification Versions

#### Version `v1`

The type specific specification fields are:

- **`baseUrl`** *string*

  OCI repository reference

- **`legacyTypes`** (optional) *bool*

  OCI repository requires docker legacy mime types for oci
  image manifests. (automatically enabled for docker.io)


### Go Bindings

The go binding can be found [here](type.go)
